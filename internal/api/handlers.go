package api

import (
	"encoding/json"
	"errors"
	"groceries/internal/grocery"
	"groceries/internal/store"
	"net/http"

	sqlite "modernc.org/sqlite"
)

// SQLITE_CONSTRAINT_FOREIGNKEY (extended result code) — surfaced when a
// SetNeeded request references a UUID that doesn't exist in the groceries
// table yet. Used by the offline queue to distinguish "drop this op" from
// "retry later".
const sqliteForeignKeyConstraint = 787

// ListAllGroceries handles GET /groceries
func ListAllGroceries(w http.ResponseWriter, r *http.Request) {
	groceries, err := store.ListAll()
	if err != nil {
		http.Error(w, "Failed to load groceries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groceries)
}

// CreateGrocery handles POST /groceries/create.
//
// Accepts either {"name":"…"} (server allocates a UUID) or
// {"uuid":"…","name":"…"} (client owns the UUID — used by the offline queue
// to replay creates idempotently). Responds 201 on a fresh insert and 200 if
// a row with the same UUID already exists, with the same JSON body in both.
func CreateGrocery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var g grocery.Grocery
	if req.UUID == "" {
		g = *grocery.NewGrocery(req.Name)
	} else {
		g = grocery.Grocery{UUID: grocery.GroceryUUID(req.UUID), Name: req.Name}
	}

	saved, created, err := store.CreateOrGet(g)
	if err != nil {
		http.Error(w, "Failed to create grocery", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(saved)
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

// SetNeededGrocery handles PUT /groceries/{uuid}/needed.
//
// Body: {"needed": true|false}. Idempotent. Returns 409 if the UUID isn't yet
// in the groceries table (FK violation) — clients should drop the op rather
// than retry.
func SetNeededGrocery(w http.ResponseWriter, r *http.Request) {
	uuidString := r.PathValue("uuid")
	if uuidString == "" {
		http.Error(w, "Missing grocery UUID", http.StatusBadRequest)
		return
	}

	var req struct {
		Needed bool `json:"needed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := store.SetNeeded(grocery.GroceryUUID(uuidString), req.Needed); err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqliteForeignKeyConstraint {
			http.Error(w, "Unknown grocery UUID", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to set needed state", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
