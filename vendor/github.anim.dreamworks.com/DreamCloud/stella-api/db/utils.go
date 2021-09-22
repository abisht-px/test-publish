package db

import (
	"errors"

	"github.anim.dreamworks.com/golang/logging"
	"github.com/jinzhu/gorm"
)

func createDB(db *gorm.DB, params map[string]string) error {
	dbName := params[DBNameKey]
	if dbName == "" {
		return errors.New("Database name is required to create new database. Set via: " + DBNameKey)
	}
	logging.Infoln("Creating database:", dbName)

	db = db.Exec("CREATE DATABASE " + dbName)

	return db.Error
}

func dropDB(db *gorm.DB, params map[string]string) error {
	dbName := params[DBNameKey]
	if dbName == "" {
		return errors.New("Database name is required to drop new database. Set via: " + DBNameKey)
	}
	logging.Infoln("Dropping database:", dbName)

	db = db.Exec("DROP DATABASE IF EXISTS " + dbName)

	return db.Error
}

func copyParamsAndChangeDbName(originalParams map[string]string, newDbName string) map[string]string {
	newParams := make(map[string]string)
	for k, v := range originalParams {
		newParams[k] = v
	}
	newParams[DBNameKey] = newDbName
	return newParams
}
