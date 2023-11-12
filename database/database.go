package database

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
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

type DB struct {
	*goqu.TxDatabase
}

var connection *Connection

// Establishes and returns a connection to the database. If a connection
// is already established, it is reused.
func GetConnection(driver, connectionUrl string) (*Connection, error) {
	if connection == nil {
		conn, err := sql.Open(driver, connectionUrl)
		if err != nil {
			return nil, err
		}
		connection = &Connection{driver, conn}
	}
	return connection, nil
}
