package repository

// Fetches the inventory of a company
func GetInventory(companyId int64) ([]any, error) {
	db, err := GetConnection()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS inventory (id INT PRIMARY KEY)")
	if err != nil {
		return nil, err
	}
	return []any{5}, nil
}

// Get the stock of a company's resource, grouping by quality
func GetStock(companyId, resourceId int64) ([]any, error) {
	return nil, nil
}
