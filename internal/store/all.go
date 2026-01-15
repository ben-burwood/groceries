package store

import (
	"encoding/json"
	"groceries/internal/grocery"
	"os"
	"sync"
)

const AllGroceryStoreFile = "store/all.json"

var allMu sync.Mutex

// loadAllGroceries reads the groceries from the JSON file
func loadAllGroceries() ([]grocery.Grocery, error) {
	file, err := os.Open(AllGroceryStoreFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []grocery.Grocery{}, nil // treat as empty list if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	var groceries []grocery.Grocery
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&groceries); err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return groceries, nil
}

// saveAllGroceries writes the groceries to the JSON file
func saveAllGroceries(groceries []grocery.Grocery) error {
	if err := os.MkdirAll("store", os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(AllGroceryStoreFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(groceries)
}

// ListAll returns all groceries.
func ListAll() ([]grocery.Grocery, error) {
	allMu.Lock()
	defer allMu.Unlock()
	return loadAllGroceries()
}

// Create adds a new grocery with the given name.
func Create(newGrocery grocery.Grocery) (*grocery.Grocery, error) {
	allMu.Lock()
	defer allMu.Unlock()

	groceries, err := loadAllGroceries()
	if err != nil {
		return nil, err
	}

	groceries = append(groceries, newGrocery)

	if err := saveAllGroceries(groceries); err != nil {
		return nil, err
	}
	return &newGrocery, nil
}
