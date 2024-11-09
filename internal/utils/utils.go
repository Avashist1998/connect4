package utils

import (
	"4connect/internal/game"
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
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

func GenerateMatchId(datastore map[string]*game.Game) string {
	var id = GenerateId(6)
	_, ok := datastore[id]
	for ok {
		id = GenerateId(6)
		_, ok = datastore[id]
	}
	return id
}

func GenerateBoardHTML(board []int) string {
	var html string
	html += "<div class='board'>\n"
	for i := 0; i < 7; i += 1 {
		html += "<div class='col' id='col-" + strconv.Itoa(i) + "'>\n"
		for j := 5; j > -1; j -= 1 {
			index := (j * 7) + i
			if board[index] == 1 {
				html += "<div class='cell player1' id='cell-" + strconv.Itoa(index) + "'>\n </div>\n"
			} else if board[index] == -1 {
				html += "<div class='cell player2' id='cell-" + strconv.Itoa(index) + "'>\n </div>\n"
			} else {
				html += "<div class='cell empty' id='cell-" + strconv.Itoa(index) + "'>\n </div>\n"
			}
		}
		html += "</div>\n"
	}
	html += "</div>\n"
	return html
}

func GenerateNewGameHTML(match *game.Game) string {

	html := "<div class='gameOver'>"
	if game.GetWinner(match) == "" {
		html += "<h2>Result: Draw</h2>\n"
		html += "<button class='resetGame'>Restart</button>\n"
		html += "<button class='newGame'>New Game</button>\n"
		html += "</div>\n"
		return html
	}
	winner := game.GetWinner(match)
	html += "<h2>Winner: " + winner + "</h2>\n"
	html += "<button class='resetGame'>Restart</button>\n"
	html += "<button class='newGame'>New Game</button>\n"
	html += "</div>\n"
	return html
}

func ReturnJson(w http.ResponseWriter, response map[string]string, status int) {
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
