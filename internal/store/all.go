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

// CreateOrGet inserts the grocery if its UUID is new, otherwise returns the
// existing row. The bool reports whether a new row was created.
func CreateOrGet(newGrocery grocery.Grocery) (*grocery.Grocery, bool, error) {
	res, err := db.Exec(
		`INSERT INTO groceries (uuid, name) VALUES (?, ?) ON CONFLICT(uuid) DO NOTHING`,
		newGrocery.UUID, newGrocery.Name,
	)
	if err != nil {
		return nil, false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	if affected == 1 {
		return &newGrocery, true, nil
	}

	var existing grocery.Grocery
	existing.UUID = newGrocery.UUID
	if err := db.QueryRow(`SELECT name FROM groceries WHERE uuid = ?`, newGrocery.UUID).Scan(&existing.Name); err != nil {
		return nil, false, err
	}
	return &existing, false, nil
}
