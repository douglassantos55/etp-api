package warehouse

import "github.com/doug-martin/goqu/v9"

func CreateResourcesTable(db *goqu.Database) {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS resources (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(255) NOT NULL,
            image VARCHAR(255)
        )
    `)

	if err != nil {
		panic(err)
	}
}

func RollbackResourcesTable(db *goqu.Database) {
	if _, err := db.Exec("DROP TABLE IF EXISTS resources"); err != nil {
		panic(err)
	}
}

func CreateInventoriesTable(db *goqu.Database) {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS inventories (
            resource_id INTEGER,
            company_id INTEGER,
            quantity INTEGER UNSIGNED,
            quality TINYINT UNSIGNED,
            sourcing_cost DECIMAL(10,2),
            PRIMARY KEY (resource_id, company_id, quality)
        )
    `)

	if err != nil {
		panic(err)
	}
}

func RollbackInventoriesTable(db *goqu.Database) {
	if _, err := db.Exec("DROP TABLE IF EXISTS inventories"); err != nil {
		panic(err)
	}
}
