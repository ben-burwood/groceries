package main

import (
	"groceries/internal/api"
	"net/http"
)

func main() {
	webMux := http.NewServeMux()
	webMux.HandleFunc("GET /groceries", api.ListAllGroceries)
	webMux.HandleFunc("POST /groceries/create", api.CreateGrocery)
	webMux.HandleFunc("GET /groceries/needed", api.ListNeededGroceries)
	webMux.HandleFunc("PUT /groceries/{uuid}/needed", api.ToggleNeededGrocery)
	// Serve Static Frontend
	webMux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	// Start web server on 8080
	http.ListenAndServe("[::]:8080", api.CORSMiddleware(webMux))
}
