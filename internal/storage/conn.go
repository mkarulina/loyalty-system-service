package storage

import (
	"database/sql"
	"github.com/spf13/viper"
	"log"
)

func initDB() *sql.DB {
	db, err := sql.Open("pgx", viper.GetString("DATABASE_URI"))
	if err != nil {
		log.Fatal(err)
	}

	return db
}
