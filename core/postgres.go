// core/postgres.go
package core

import (
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

// DATABASE URL "postgres://postgres:password@localhost:5432/noble"
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sqlx.Open("pgx", dataSourceName)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	// Настройки пула соединений
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected successfully (sqlx + pgx)")
	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
