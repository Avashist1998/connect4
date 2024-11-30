package utils

import (
	"4connect/internal/models"
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
)

func GenerateId(size int) string {
	choices := "abcdefgijklmnopqrstuvwxyz123456789"
	id := make([]byte, size)
	for i := range id {
		id[i] = choices[rand.Intn(len(choices))]
	}
	return string(id)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/" + tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GenerateMatchId(datastore map[string]*models.MatchSession) string {
	var id = GenerateId(6)
	_, ok := datastore[id]
	for ok {
		id = GenerateId(6)
		_, ok = datastore[id]
	}
	return id
}

func ReturnJson(w http.ResponseWriter, response map[string]interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func AbsInt64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
