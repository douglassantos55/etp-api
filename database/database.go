package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	POSTGRES = "postgres"
)

type Connection struct {
	Driver string
	DB     *sql.DB
}

// Establishes and returns a connection to the database. If a connection
// is already established, it is reused.
func GetConnection(driver, connectionUrl string) (*Connection, error) {
	conn, err := sql.Open(driver, connectionUrl)
	if err != nil {
		return nil, err
	}
	return &Connection{driver, conn}, nil
}
