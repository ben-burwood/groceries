package store

import "groceries/internal/grocery"

// ListAll returns all groceries ordered by name.
func ListAll() ([]grocery.Grocery, error) {
	rows, err := db.Query(`SELECT uuid, name FROM groceries ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groceries := []grocery.Grocery{}
	for rows.Next() {
		var g grocery.Grocery
		if err := rows.Scan(&g.UUID, &g.Name); err != nil {
			return nil, err
		}
		groceries = append(groceries, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groceries, nil
}

// Create adds a new grocery.
func Create(newGrocery grocery.Grocery) (*grocery.Grocery, error) {
	_, err := db.Exec(
		`INSERT INTO groceries (uuid, name) VALUES (?, ?)`,
		newGrocery.UUID, newGrocery.Name,
	)
	if err != nil {
		return nil, err
	}
	return &newGrocery, nil
}
