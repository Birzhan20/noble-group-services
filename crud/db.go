package crud

import "github.com/jmoiron/sqlx"

// SetDB sets the database connection for the crud package.
func SetDB(database *sqlx.DB) {
	db = database
}
