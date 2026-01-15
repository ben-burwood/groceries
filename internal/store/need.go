package store

import (
	"encoding/json"
	"groceries/internal/grocery"
	"os"
	"sync"
)

const GroceryNeedStoreFile = "store/need.json"

var needMu sync.Mutex

// loadNeededGroceries reads the groceries from the JSON file
func loadNeededGroceries() ([]grocery.GroceryUUID, error) {
	file, err := os.Open(GroceryNeedStoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []grocery.GroceryUUID{}, nil // treat as empty list if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	var groceries []grocery.GroceryUUID
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&groceries); err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return groceries, nil
}

// saveNeededGroceries writes the groceries to the JSON file
func saveNeededGroceries(groceries []grocery.GroceryUUID) error {
	if err := os.MkdirAll("store", os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(GroceryNeedStoreFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(groceries)
}

// ListNeeded returns all needed groceries.
func ListNeeded() ([]grocery.GroceryUUID, error) {
	needMu.Lock()
	defer needMu.Unlock()
	return loadNeededGroceries()
}

// ToggleNeeded toggles a grocery as needed.
func ToggleNeeded(groceryUUID grocery.GroceryUUID) error {
	needMu.Lock()
	defer needMu.Unlock()

	groceries, err := loadNeededGroceries()
	if err != nil {
		return err
	}

	for i, uuid := range groceries {
		if uuid == groceryUUID {
			// Grocery was needed, remove from list to mark as got
			groceries = append(groceries[:i], groceries[i+1:]...)
			return saveNeededGroceries(groceries)
		}
	}
	groceries = append(groceries, groceryUUID)
	return saveNeededGroceries(groceries)
}
