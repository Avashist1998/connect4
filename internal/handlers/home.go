package handlers

import (
	"4connect/internal/utils"
	"net/http"
	"strings"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		utils.RenderTemplate(w, "index.html", nil)
	case http.MethodPost:
		HandleMakeMatch(w, r)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		handleRoot(w, r)
	} else {
		HandleMatch(w, r, path)
	}
}
