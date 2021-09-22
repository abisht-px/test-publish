package db

import (
	"github.anim.dreamworks.com/golang/logging"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"
)

var db *gorm.DB
var dbErr error

func Connect() (*gorm.DB, error) {
	if db == nil {
		dbErr = setupDatabase()
	}

	return db, dbErr
}

func Close() {
	if db != nil {
		logging.Infoln("Closing DB connection.")
		db.Close()
	}
}

func setupDatabase() error {
	env := utils.EnvOrDefault("GO_ENV", "dev")
	params := DBParams(env)

	if env == "test" {
		err := dropCreateDB(params)
		if err != nil {
			return err
		}
	}

	return initDB(params)
}

func dropCreateDB(params map[string]string) error {
	postgresDbParams := copyParamsAndChangeDbName(params, "postgres")
	var err error

	// The default DB connection cannot modify its working database.
	// Therefore, I create a short-lived connection that uses the always-present "postgres" DB.
	postgresDb, err := gorm.Open("postgres", dbConfig(postgresDbParams))
	if err != nil {
		return err
	}
	defer postgresDb.Close()
	err = dropDB(postgresDb, params)
	if err != nil {
		return err
	}
	err = createDB(postgresDb, params)
	if err != nil {
		return err
	}
	return nil
}

func initDB(params map[string]string) error {
	db, dbErr = gorm.Open("postgres", dbConfig(params))

	if dbErr == nil {
		runMigrations(db)
	}
	return dbErr
}

func IsDatabaseOK() bool {
	return dbErr == nil
}
