package grocery

import (
	"github.com/google/uuid"
)

type GroceryUUID string

type Grocery struct {
	UUID GroceryUUID `json:"uuid"`
	Name string      `json:"name"`
}

// NewGrocery - Constructor to create a new Grocery with a unique UUID.
func NewGrocery(name string) *Grocery {
	return &Grocery{
		UUID: GroceryUUID(uuid.NewString()),
		Name: name,
	}
}
