package handlers

import (
	"4connect/internal/game"
	"4connect/internal/models"
	"4connect/internal/utils"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

var datastore = make(map[string]*game.Game)

func MatchHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		var data models.MatchData
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Log the body as a string
		log.Printf("Request Body: %s", string(bodyBytes))
		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newGame := game.NewGame(data.Player1, data.Player2)
		id := utils.GenerateMatchId(datastore)
		datastore[id] = newGame
		response := map[string]string{
			"match_id": id,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func MatchPlayHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/match/")
	matchID := path
	match, ok := datastore[matchID]
	// Check if the match exists in the datastore
	if !ok {
		response := map[string]string{
			"message": "invalid match id",
		}
		utils.ReturnJson(w, response, http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		newGameHTML := ""
		boardHTML := utils.GenerateBoardHTML(game.GetBoard(match))

		if game.IsGameOver(match) {
			newGameHTML = utils.GenerateNewGameHTML(match)
		}
		var data = models.MatchPageData{
			Player1:     match.Player1,
			Player2:     match.Player2,
			CurrPlayer:  game.GetCurrPlayer(match),
			BoardHTML:   template.HTML(boardHTML),
			NewGameHTML: template.HTML(newGameHTML),
		}
		utils.RenderTemplate(w, "match.html", data)

	case http.MethodPost:
		var data models.MoveData
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			response := map[string]string{
				"message": "invalid body",
			}
			utils.ReturnJson(w, response, http.StatusBadRequest)
			return
		}

		log.Printf("Request Body: %s", string(bodyBytes))

		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			response := map[string]string{
				"message": "Invalid JSON in request",
			}
			utils.ReturnJson(w, response, http.StatusBadRequest)
			return
		}
		err = game.MakeMove(match, data.Player, data.Move)
		if err != nil {
			response := map[string]string{
				"message": "invalid move or game over",
			}
			utils.ReturnJson(w, response, http.StatusBadGateway)
			return
		}
		response := map[string]string{
			"message": "Move successfully updated",
		}
		utils.ReturnJson(w, response, http.StatusOK)

	case http.MethodDelete:
		newGame := game.NewGame(match.Player1, match.Player2)
		datastore[matchID] = newGame
		response := map[string]string{
			"message": "Game has been reset",
		}
		utils.ReturnJson(w, response, http.StatusOK)

	default:
		response := map[string]string{
			"message": "Method not allowed",
		}
		utils.ReturnJson(w, response, http.StatusMethodNotAllowed)
	}
}
