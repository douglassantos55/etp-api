package database

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	POSTGRES = "postgres"
)

var (
	db *sql.DB

	builder *goqu.Database
)

// Establishes and returns a connection to the database. If a connection
// is already established, it is reused.
func GetConnection(driver, connectionUrl string) (*sql.DB, error) {
	if db == nil {
		conn, err := sql.Open(driver, connectionUrl)
		if err != nil {
			return nil, err
		}
		db = conn
	}
	return db, nil
}

// Creates the query builder
func GetBuilder(driver, connectionUrl string) (*goqu.Database, error) {
	if builder == nil {
		db, err := GetConnection(driver, connectionUrl)
		if err != nil {
			return nil, err
		}
		builder = goqu.New(driver, db)
	}
	return builder, nil
}
