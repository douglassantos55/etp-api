package repository

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var db *goqu.Database

// Establishes and returns a connection to the database. If a connection
// is already established, it is reused.
func GetConnection() (*goqu.Database, error) {
	if db == nil {
		conn, err := sql.Open("sqlite3", "database.sqlite")
		if err != nil {
			return nil, err
		}
		db = goqu.New("sqlite3", conn)
	}
	return db, nil
}
