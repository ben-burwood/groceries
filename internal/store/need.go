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

// SetNeeded marks a grocery as needed (true) or not needed (false). Both
// directions are idempotent. A FK violation propagates back to the caller so
// the API layer can surface it as 409 — that happens when the UUID isn't yet
// in the groceries table (e.g. a queued setNeeded sync arrives before its
// matching create).
func SetNeeded(groceryUUID grocery.GroceryUUID, needed bool) error {
	if needed {
		_, err := db.Exec(
			`INSERT INTO needed_groceries (uuid) VALUES (?) ON CONFLICT(uuid) DO NOTHING`,
			groceryUUID,
		)
		return err
	}
	_, err := db.Exec(`DELETE FROM needed_groceries WHERE uuid = ?`, groceryUUID)
	return err
}
