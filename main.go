package main

import (
	"groceries/internal/api"
	"groceries/internal/store"
	"log"
	"net/http"
)

func main() {
	if err := store.Init(); err != nil {
		log.Fatalf("store init: %v", err)
	}
	defer store.Close()

	webMux := http.NewServeMux()
	webMux.HandleFunc("GET /groceries", api.ListAllGroceries)
	webMux.HandleFunc("POST /groceries/create", api.CreateGrocery)
	webMux.HandleFunc("GET /groceries/needed", api.ListNeededGroceries)
	webMux.HandleFunc("PUT /groceries/{uuid}/needed", api.SetNeededGrocery)
	// Serve Static Frontend
	webMux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	// Start web server on 8080
	log.Fatal(http.ListenAndServe("[::]:8080", api.CORSMiddleware(webMux)))
}
