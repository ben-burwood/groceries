package store

import "groceries/internal/grocery"

func ListNeeded() ([]grocery.GroceryUUID, error) {
	rows, err := db.Query(`SELECT uuid FROM needed_groceries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	needed := []grocery.GroceryUUID{}
	for rows.Next() {
		var u grocery.GroceryUUID
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		needed = append(needed, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return needed, nil
}

func ToggleNeeded(groceryUUID grocery.GroceryUUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM needed_groceries WHERE uuid = ?`, groceryUUID)
	if err != nil {
		return err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		if _, err := tx.Exec(`INSERT INTO needed_groceries (uuid) VALUES (?)`, groceryUUID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
