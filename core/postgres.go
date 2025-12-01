package core

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DB is the database connection.
var DB *sql.DB

// InitDB initializes the database connection.
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	fmt.Println("Successfully connected to the database!")
	return nil
}
