package main

import (
	"4connect/internal/handlers"
	"4connect/internal/utils"
	"log"
	"net/http"
	"strings"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		switch r.Method {
		case http.MethodGet:
			utils.RenderTemplate(w, "index.html", nil)
		case http.MethodPost:
			handlers.MatchHandler(w, r)
		}
	} else {
		handlers.MatchPlayHandler(w, r, path)
	}
}

func main() {

	port := "0.0.0.0:9080"
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws/live", handlers.LiveWebSocketHandler)
	log.Println("Starting server on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
