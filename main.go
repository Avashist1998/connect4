package main

import (
	"4connect/internal/handlers"
	"4connect/internal/utils"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "index.html", nil)
}

func main() {

	port := ":9080"
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/match", handlers.MatchHandler)
	http.HandleFunc("/match/", handlers.MatchPlayHandler)
	log.Println("Starting server on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
