package datastore

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // need to import because database/sql only contains specs not implementation
)

type sslMode string

const (
	postgresDB         = "postgres"
	VerifyCA   sslMode = "verify-ca"
	Require    sslMode = "require"
	Disable    sslMode = "disable"
)

type DBDriver interface {
	CheckConnection() error
}

type postgresDriver struct {
	host        string
	port        int
	dbName      string
	password    string
	userName    string
	sslMode     sslMode
	sslRootCert string
}

func NewPostgresDriver(port int, host, dbName, userName, password, sslRootCerts string, sslMode sslMode) DBDriver {
	return postgresDriver{
		host:        host,
		dbName:      dbName,
		password:    password,
		port:        port,
		userName:    userName,
		sslMode:     sslMode,
		sslRootCert: sslRootCerts,
	}
}

func (driver postgresDriver) CheckConnection() error {
	psqlConStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", driver.host, driver.port, driver.userName, driver.password, driver.dbName, driver.sslMode)
	if driver.sslMode == VerifyCA {
		psqlConStr = fmt.Sprintf("%s sslrootcert=%s", psqlConStr, driver.sslRootCert)
	}
	db, err := sql.Open(postgresDB, psqlConStr)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}
