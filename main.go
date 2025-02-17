package main

import (
	"4connect/internal/handlers"

	"log"
	"net/http"
)

func main() {

	port := "0.0.0.0:9080"
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handlers.HomeHandler)
	// http.HandleFunc("/ws/live", handlers.LiveWebSocketHandler)
	http.HandleFunc("/ws/live", handlers.LiveWebSocketHandlerV2)
	http.HandleFunc("/ws/lobby", handlers.LobbyWebSocketHandler)
	log.Printf("Starting server on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
