package db

import (
	"fmt"
	"strings"

	"github.anim.dreamworks.com/golang/logging"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"
)

const (
	HostKey   = "DATABASE_URL"
	PortKey   = "DATABASE_PORT"
	DBNameKey = "DATABASE_NAME"
	UserKey   = "DATABASE_USER"
	PassKey   = "DATABASE_PASSWORD"
	SSLKey    = "DATABASE_SSL_MODE"
)

var testDefaults = map[string]string{
	HostKey:   "localhost",
	PortKey:   "5432",
	DBNameKey: "stella_test",
	UserKey:   "stella",
	PassKey:   "stella",
	SSLKey:    "disable",
}

var devDefaults = map[string]string{
	HostKey:   "localhost",
	PortKey:   "5432",
	DBNameKey: "stella_dev",
	UserKey:   "stella",
	PassKey:   "stella",
	SSLKey:    "disable",
}

var genDefaults = map[string]string{
	PortKey: "5432",
	SSLKey:  "disable",
}

func dbConfig(p map[string]string) string {
	configStr := "host=%v port=%v dbname=%v user=%v password=%v sslmode=%v"
	return fmt.Sprintf(configStr, p[HostKey], p[PortKey], p[DBNameKey], p[UserKey], p[PassKey], p[SSLKey])
}

func DBParams(env string) map[string]string {
	params := make(map[string]string)
	envPrefix := strings.ToUpper(env) + "_"

	var defaults map[string]string

	switch env {
	case "test":
		defaults = testDefaults
	case "dev":
		defaults = devDefaults
	default:
		defaults = genDefaults
	}

	host := utils.EnvOrDefault(envPrefix+HostKey, defaults[HostKey])
	if host == "" {
		logging.Fatalln(envPrefix+HostKey, "environment variable was not set (required). Exiting.")
	}

	dbName := utils.EnvOrDefault(envPrefix+DBNameKey, defaults[DBNameKey])
	if dbName == "" {
		logging.Fatalln(envPrefix+DBNameKey, "environment variable was not set (required). Exiting.")
	}

	user := utils.EnvOrDefault(envPrefix+UserKey, defaults[UserKey])
	if user == "" {
		logging.Fatalln(envPrefix+UserKey, "environment variable was not set (required). Exiting.")
	}

	pass := utils.EnvOrDefault(envPrefix+PassKey, defaults[PassKey])
	if pass == "" {
		logging.Fatalln(envPrefix+PassKey, "environment variable was not set (required). Exiting.")
	}

	port := utils.EnvOrDefault(envPrefix+PortKey, defaults[PortKey])
	if port == "" {
		logging.Fatalln(envPrefix+PassKey, "environment variable was not set (required). Exiting.")
	}

	ssl := utils.EnvOrDefault(envPrefix+SSLKey, defaults[SSLKey])
	if ssl == "" {
		logging.Fatalln(envPrefix+SSLKey, "environment variable was not set (required). Exiting.")
	}

	params[HostKey] = host
	params[PortKey] = port
	params[DBNameKey] = dbName
	params[UserKey] = user
	params[PassKey] = pass
	params[SSLKey] = ssl

	return params
}
