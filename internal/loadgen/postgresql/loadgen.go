package postgresql

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	keyPrefix = "key-"
)

type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

type Config struct {
	User        string
	Password    string
	Host        string
	Port        string
	DBName      string
	DBTableName string
	Count       int
	Logger      Logger
}

func (c Config) Validate() error {
	switch {
	case c.User == "":
		return fmt.Errorf("user should be set")
	case c.Password == "":
		return fmt.Errorf("password should be set")
	case c.Host == "":
		return fmt.Errorf("host should be set")
	case c.DBName == "":
		return fmt.Errorf("database name should be set")
	}
	return nil
}

type Loader struct {
	cfg Config

	stopCh chan struct{}
	conn   *pgx.Conn
}

func New(cfg Config) (*Loader, error) {
	if cfg.DBName == "" {
		cfg.DBName = "postgres"
	}
	if cfg.Count < 1 {
		cfg.Count = 100
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Loader{
		stopCh: make(chan struct{}),
		cfg:    cfg,
	}, nil
}

func (l *Loader) Name() string {
	return "PostgreSQL"
}

func (l *Loader) Run(ctx context.Context) (err error) {
	if err := l.init(ctx); err != nil {
		return err
	}
	defer func() {
		err = l.shutdown(ctx)
	}()
	l.runOnce(ctx)
	return nil
}

func (l *Loader) RunWithInterval(ctx context.Context, interval time.Duration) error {
	if err := l.init(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return l.shutdown(ctx)
		default:
			l.runOnce(ctx)
			time.Sleep(interval)
		}
	}
}

func (l *Loader) init(ctx context.Context) error {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		l.cfg.User, l.cfg.Password, l.cfg.Host, l.cfg.Port, l.cfg.DBName)
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %s", err)
	}
	l.conn = conn

	l.cfg.DBTableName = l.getTableName(l.cfg.DBTableName)
	if err := l.ensureTable(ctx); err != nil {
		return fmt.Errorf("create test table(s): %s", err)
	}
	l.printf("Use %s table (%s database).", l.cfg.DBTableName, l.cfg.DBName)
	return nil
}

func (l *Loader) shutdown(ctx context.Context) error {
	defer l.conn.Close(ctx)

	l.print("Stopping the test.")
	l.printf("Removing the %s test table...\n", l.cfg.DBTableName)
	return l.dropTable(ctx)
}

func (l *Loader) runOnce(ctx context.Context) {
	l.runInserts(ctx, l.conn, l.cfg.DBTableName, l.cfg.Count)
	l.runQueries(ctx, l.conn, l.cfg.DBTableName, l.cfg.Count)
	l.runUpdates(ctx, l.conn, l.cfg.DBTableName, l.cfg.Count)
	l.runDeletions(ctx, l.conn, l.cfg.DBTableName, l.cfg.Count)
	l.runRollbacks(ctx, l.conn, l.cfg.DBTableName, l.cfg.Count)
}

func (l *Loader) runInserts(ctx context.Context, conn *pgx.Conn, tableName string, count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		value := time.Now().String()
		query := fmt.Sprintf("INSERT INTO %s VALUES($1, $2)", tableName)
		_, err := conn.Exec(ctx, query, key, value)
		if err != nil {
			l.printf("ERROR: run inserts: key %s: %s.", key, err)
		}
	}
	stop := time.Now()
	l.printf("%d inserts done in %v.", count, stop.Sub(start))
}

func (l *Loader) runQueries(ctx context.Context, conn *pgx.Conn, tableName string, count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		var resValue string
		query := fmt.Sprintf("SELECT value FROM %s WHERE key=$1", tableName)
		err := conn.QueryRow(ctx, query, key).
			Scan(&resValue)
		if err != nil {
			l.printf("ERROR: run queries: key %s: %s.", key, err)
		}
	}
	stop := time.Now()
	l.printf("%d queries done in %v.", count, stop.Sub(start))
}

func (l *Loader) runUpdates(ctx context.Context, conn *pgx.Conn, tableName string, count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		value := time.Now().String()
		query := fmt.Sprintf("UPDATE %s SET value=$1 WHERE key=$2", tableName)
		_, err := conn.Exec(ctx, query, value, key)
		if err != nil {
			l.printf("ERROR: run updates: key %s: %s.", key, err)
		}
	}
	stop := time.Now()
	l.printf("%d updates done in %v.", count, stop.Sub(start))
}

func (l *Loader) runDeletions(ctx context.Context, conn *pgx.Conn, tableName string, count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		query := fmt.Sprintf("DELETE FROM %s WHERE key=$1", tableName)
		_, err := conn.Exec(ctx, query, key)
		if err != nil {
			l.printf("ERROR: run deletions: key %s: %s.", key, err)
		}
	}
	stop := time.Now()
	l.printf("%d deletes done in %v.", count, stop.Sub(start))
}

func (l *Loader) runRollbacks(ctx context.Context, conn *pgx.Conn, tableName string, count int) {
	start := time.Now()
	for i := 0; i < count; i++ {
		tx, err := conn.Begin(ctx)
		if err != nil {
			l.printf("ERROR: create transaction: %s.", err)
		}
		key := keyPrefix + strconv.Itoa(i)
		value := time.Now().String()
		query := fmt.Sprintf("INSERT INTO %s VALUES($1, $2)", tableName)
		_, err = tx.Exec(ctx, query, key, value)
		if err != nil {
			l.printf("ERROR: run transaction: insert: key %s: %s.", key, err)
		}
		err = tx.Rollback(ctx)

		if err != nil {
			l.printf("ERROR: transaction rollback: %s.", err)
		}
	}
	stop := time.Now()
	l.printf("%d rollbacks done in %v.", count, stop.Sub(start))
}

func (l *Loader) ensureTable(ctx context.Context) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  key varchar(45) NOT NULL,
  value varchar(450) NOT NULL,
  PRIMARY KEY (key)
)`, l.cfg.DBTableName)

	_, err := l.conn.Exec(ctx, query)
	return err
}

func (l *Loader) dropTable(ctx context.Context) error {
	query := fmt.Sprintf("DROP TABLE %s", l.cfg.DBTableName)
	_, err := l.conn.Exec(ctx, query)
	return err
}

func (l *Loader) print(message string) {
	if l.cfg.Logger != nil {
		l.cfg.Logger.Print(message)
	}
}

func (l *Loader) printf(format string, args ...interface{}) {
	if l.cfg.Logger != nil {
		l.cfg.Logger.Printf(format, args...)
	}
}

func (l *Loader) getTableName(tableName string) string {
	if tableName != "" {
		return tableName
	}
	tableName = "testtable"
	tableName = strings.Replace(tableName, "-", "", -1)
	return fmt.Sprintf("%s%d", tableName, time.Now().Nanosecond())
}
