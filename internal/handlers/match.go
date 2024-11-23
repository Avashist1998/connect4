package handlers

import (
	"4connect/internal/game"
	"4connect/internal/models"
	"4connect/internal/store"
	"4connect/internal/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func MatchHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		var data models.MatchRequestData
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
		datastore := store.GetDataStore()
		id := utils.GenerateMatchId(datastore)
		datastore[id] = &models.Match{
			GameType: data.GameType,
			Game:     game.NewGame(data.Player1, data.Player2)}
		response := map[string]interface{}{
			"match_id": id,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func MatchPlayHandler(w http.ResponseWriter, r *http.Request, matchID string) {

	datastore := store.GetDataStore()
	data, ok := datastore[matchID]
	// Check if the match exists in the datastore
	if !ok {
		response := map[string]interface{}{
			"message": "invalid match id",
		}
		utils.ReturnJson(w, response, http.StatusNotFound)
		return
	}

	if data.GameType == "local" {
		MatchLocalPlayHandler(w, r, matchID, data)
	} else if data.GameType == "live" {
		var data = models.LivePageData{MatchID: matchID}
		utils.RenderTemplate(w, "live.html", data)
	}
}

func MatchLocalPlayHandler(w http.ResponseWriter, r *http.Request, matchID string, matchData *models.Match) {
	match := matchData.Game
	switch r.Method {
	case http.MethodGet:
		var data = models.MatchPageData{
			Player1:    match.Player1,
			Player2:    match.Player2,
			CurrPlayer: game.GetCurrPlayer(match),
		}
		utils.RenderTemplate(w, "match.html", data)

	case http.MethodPost:
		var data models.MoveData
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			response := map[string]interface{}{
				"message": "invalid body",
			}
			utils.ReturnJson(w, response, http.StatusBadRequest)
			return
		}

		log.Printf("Request Body: %s", string(bodyBytes))

		err = json.Unmarshal(bodyBytes, &data)
		if err != nil {
			response := map[string]interface{}{
				"message": "Invalid JSON in request",
			}
			utils.ReturnJson(w, response, http.StatusBadRequest)
			return
		}
		err = game.MakeMove(match, data.Player, data.Move)
		if err != nil {
			response := map[string]interface{}{
				"message": "invalid move or game over",
			}
			utils.ReturnJson(w, response, http.StatusBadGateway)
			return
		}
		response := map[string]interface{}{
			"message":    "Move successfully updated",
			"board":      game.GetBoard(match),
			"currPlayer": game.GetCurrPlayer(match),
			"winner":     "",
		}

		if game.IsGameOver(match) {
			response = map[string]interface{}{
				"message":    "Game Over",
				"board":      game.GetBoard(match),
				"currPlayer": game.GetCurrPlayer(match),
				"winner":     game.GetWinner(match),
			}
		}
		utils.ReturnJson(w, response, http.StatusOK)

	case http.MethodDelete:
		matchData.Game = game.NewGame(match.Player1, match.Player2)
		response := map[string]interface{}{
			"message": "Game has been reset",
		}
		utils.ReturnJson(w, response, http.StatusOK)

	default:
		response := map[string]interface{}{
			"message": "Method not allowed",
		}
		utils.ReturnJson(w, response, http.StatusMethodNotAllowed)
	}

}
