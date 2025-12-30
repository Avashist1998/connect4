package main

import (
	"4connect/internal/handlers"

	"log"
	"net/http"
)

func main() {

	port := "0.0.0.0:9080"
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/session/{sessionId}", handlers.HandleSessionGet)
	mux.HandleFunc("/session", handlers.HandleSessionPost)
	mux.HandleFunc("/ws/live", handlers.LiveWebSocketHandler)

	mux.HandleFunc("/", handlers.HomeHandler)
	// http.HandleFunc("/ws/lobby", handlers.LobbyWebSocketHandler)
	log.Println("Starting server on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
