package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	c "github.com/over55/workery-cli/config"
)

func NewStorage(appCfg *c.Conf, schemaName string) *sql.DB {
	log.Println("storage postgres initializing...")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		appCfg.PostgresDB.DatabaseHost,
		appCfg.PostgresDB.DatabasePort,
		appCfg.PostgresDB.DatabaseUser,
		appCfg.PostgresDB.DatabasePassword,
		appCfg.PostgresDB.DatabaseName,
		schemaName,
	)

	log.Println("storage postgres config:", psqlInfo)

	dbInstance, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	err = dbInstance.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("storage postgres initialized successfully")
	return dbInstance
}
