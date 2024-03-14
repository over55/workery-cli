package config

import (
	"log"
	"os"
	"strconv"
)

type Conf struct {
	PostgresDB postgresDBConfig
	DB         mongoDBConfig
	AWS        awsConfig
	OldAWS     awsConfig
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

type awsConfig struct {
	AccessKey      string
	SecretKey      string
	Endpoint       string
	Region         string
	BucketName     string
	ForcePathStyle bool
}

func New() *Conf {
	var c Conf

	c.DB.URI = getEnv("WORKERY_BACKEND_DB_URI", true)
	c.DB.Name = getEnv("WORKERY_BACKEND_DB_NAME", true)

	c.PostgresDB.DatabaseHost = getEnv("WORKERY_BACKEND_DB_HOST", true)
	c.PostgresDB.DatabasePort = getEnv("WORKERY_BACKEND_DB_PORT", true)
	c.PostgresDB.DatabaseUser = getEnv("WORKERY_BACKEND_DB_USER", true)
	c.PostgresDB.DatabasePassword = getEnv("WORKERY_BACKEND_DB_PASSWORD", false)
	c.PostgresDB.DatabaseName = getEnv("WORKERY_BACKEND_DB_NAME", true)
	c.PostgresDB.DatabasePublicSchemaName = getEnv("WORKERY_BACKEND_PUBLIC_SCHEMA_NAME", true)
	c.PostgresDB.DatabaseLondonSchemaName = getEnv("WORKERY_BACKEND_LONDON_SCHEMA_NAME", true)
	c.AWS.AccessKey = getEnv("WORKERY_BACKEND_AWS_ACCESS_KEY", true)
	c.AWS.SecretKey = getEnv("WORKERY_BACKEND_AWS_SECRET_KEY", true)
	c.AWS.Endpoint = getEnv("WORKERY_BACKEND_AWS_ENDPOINT", true)
	c.AWS.Region = getEnv("WORKERY_BACKEND_AWS_REGION", true)
	c.AWS.BucketName = getEnv("WORKERY_BACKEND_AWS_BUCKET_NAME", true)
	c.AWS.ForcePathStyle = getEnvBool("WORKERY_BACKEND_AWS_S3_FORCE_PATH_STYLE", false, false)
	c.OldAWS.AccessKey = getEnv("WORKERY_BACKEND_OLD_AWS_ACCESS_KEY", true)
	c.OldAWS.SecretKey = getEnv("WORKERY_BACKEND_OLD_AWS_SECRET_KEY", true)
	c.OldAWS.Endpoint = getEnv("WORKERY_BACKEND_OLD_AWS_ENDPOINT", true)
	c.OldAWS.Region = getEnv("WORKERY_BACKEND_OLD_AWS_REGION", true)
	c.OldAWS.BucketName = getEnv("WORKERY_BACKEND_OLD_AWS_BUCKET_NAME", true)

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
