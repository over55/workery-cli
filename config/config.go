package config

import (
	"log"
	"os"
	"strconv"
)

type Conf struct {
	PostgresDB postgresDBConfig
	DB         mongoDBConfig
}

type mongoDBConfig struct {
	URI  string
	Name string
}

type postgresDBConfig struct {
	DatabaseHost             string
	DatabasePort             string
	DatabaseUser             string
	DatabasePassword         string
	DatabaseName             string
	DatabasePublicSchemaName string
	DatabaseLondonSchemaName string
}

func New() *Conf {
	var c Conf

	c.DB.URI = getEnv("WORKERY_BACKEND_DB_URI", true)
	c.DB.Name = getEnv("WORKERY_BACKEND_DB_NAME", true)

	c.PostgresDB.DatabaseHost = getEnv("WORKERY_BACKEND_DB_HOST", true)
	c.PostgresDB.DatabasePort = getEnv("WORKERY_BACKEND_DB_PORT", true)
	c.PostgresDB.DatabaseUser = getEnv("WORKERY_BACKEND_DB_USER", true)
	c.PostgresDB.DatabasePassword = getEnv("WORKERY_BACKEND_DB_PASSWORD", true)
	c.PostgresDB.DatabaseName = getEnv("WORKERY_BACKEND_DB_NAME", true)
	c.PostgresDB.DatabasePublicSchemaName = getEnv("WORKERY_BACKEND_PUBLIC_SCHEMA_NAME", true)
	c.PostgresDB.DatabaseLondonSchemaName = getEnv("WORKERY_BACKEND_LONDON_SCHEMA_NAME", true)

	return &c
}

func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("Environment variable not found: %s", key)
	}
	return value
}

func getEnvBool(key string, required bool, defaultValue bool) bool {
	valueStr := getEnv(key, required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Fatalf("Invalid boolean value for environment variable %s", key)
	}
	return value
}
