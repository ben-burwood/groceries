package api

import (
	"encoding/json"
	"groceries/internal/grocery"
	"groceries/internal/store"
	"net/http"
)

// ListAllGroceries handles GET /todos
func ListAllGroceries(w http.ResponseWriter, r *http.Request) {
	groceries, err := store.ListAll()
	if err != nil {
		http.Error(w, "Failed to load groceries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groceries)
}

// CreateGrocery handles POST /groceries/create
func CreateGrocery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	newGrocery, err := store.Create(*grocery.NewGrocery(req.Name))
	if err != nil {
		http.Error(w, "Failed to create grocery", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGrocery)
}

// ListNeededGroceries handles GET /groceries/needed
func ListNeededGroceries(w http.ResponseWriter, r *http.Request) {
	groceries, err := store.ListNeeded()
	if err != nil {
		http.Error(w, "Failed to load groceries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groceries)
}

// ToggleNeededGrocery handles PUT /groceries/{uuid}/needed
func ToggleNeededGrocery(w http.ResponseWriter, r *http.Request) {
	uuidString := r.PathValue("uuid")
	if uuidString == "" {
		http.Error(w, "Missing grocery UUID", http.StatusBadRequest)
		return
	}
	err := store.ToggleNeeded(grocery.GroceryUUID(uuidString))
	if err != nil {
		http.Error(w, "Failed to toggle grocery", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
