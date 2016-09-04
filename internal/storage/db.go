package storage

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// register postgresql driver
	_ "github.com/lib/pq"
)

// OpenDatabase opens the database and performs a ping to make sure the
// database is up.
func OpenDatabase(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection error: %s", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database error: %s", err)
	}
	return db, nil
}
