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

func HandleMakeMatch(w http.ResponseWriter, r *http.Request) {
	/*
		Handle the creation of a match in the system
		input: request of MatchRequest
		response: return the match id as response

	*/
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
	datastore := store.MatchManagerFactory()
	var matchID string = datastore.CreateGame(data.Player1, data.Player2, data.Level, data.GameType)
	response := map[string]interface{}{
		"match_id": matchID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func HandleMatch(w http.ResponseWriter, r *http.Request, matchID string) {

	datastore := store.MatchManagerFactory()
	data, err := datastore.GetMatchSession(matchID)
	if err != nil {
		response := map[string]interface{}{
			"message": "invalid match id",
		}
		utils.ReturnJson(w, response, http.StatusNotFound)
		return
	}

	if data.GameType == "local" {
		HandleLocalPlay(w, r, matchID, data)
	} else if data.GameType == "live" {
		var data = models.LivePageData{MatchID: matchID}
		utils.RenderTemplate(w, "live.html", data)
	} else {
		MatchBotPlayHandler(w, r, matchID, data)
	}
}

func MatchBotPlayHandler(w http.ResponseWriter, r *http.Request, matchID string, matchData *models.MatchSession) {
	match := matchData.Game
	switch r.Method {
	case http.MethodGet:
		var data = models.MatchPageData{
			Player1:    match.Player1,
			Player2:    match.Player2,
			CurrPlayer: match.GetCurrPlayer(),
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
		err = game.MakeMove(match, data.Slot, data.Move)
		if err != nil {
			response := map[string]interface{}{
				"message": "invalid move or game over",
			}
			utils.ReturnJson(w, response, http.StatusBadGateway)
			return
		}
		if !match.IsGameOver() {
			// TODO: We need to update the yellow so that bot can be any slot
			move := match.GetBotMove(matchData.Level, "YELLOW")
			game.MakeMove(match, "YELLOW", move)
		}
		response := map[string]interface{}{
			"message":    "Move successfully updated",
			"board":      match.GetBoard(),
			"currPlayer": match.GetCurrPlayer(),
			"currSlot":   match.GetCurrSlot(),
			"winner":     "",
		}

		if match.IsGameOver() {
			response = map[string]interface{}{
				"message":    "Game Over",
				"board":      match.GetBoard(),
				"currPlayer": match.GetCurrPlayer(),
				"winner":     match.GetWinner(),
			}
		}
		utils.ReturnJson(w, response, http.StatusOK)

	case http.MethodDelete:
		matchData.Game = game.NewGame(match.Player1, match.Player2)
		response := map[string]interface{}{
			"message": "Game has been reset",
		}
		utils.ReturnJson(w, response, http.StatusOK)
	}
}

func HandleLocalPlay(w http.ResponseWriter, r *http.Request, matchID string, matchData *models.MatchSession) {
	match := matchData.Game
	switch r.Method {
	case http.MethodGet:
		var data = models.MatchPageData{
			Player1:    match.Player1,
			Player2:    match.Player2,
			CurrSlot:   match.GetCurrSlot(),
			CurrPlayer: match.GetCurrPlayer(),
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

		err = game.MakeMove(match, data.Slot, data.Move)
		if err != nil {
			response := map[string]interface{}{
				"message": "invalid move or game over",
			}
			utils.ReturnJson(w, response, http.StatusBadGateway)
			return
		}
		response := map[string]interface{}{
			"message":    "Move successfully updated",
			"board":      match.GetBoard(),
			"currPlayer": match.GetCurrPlayer(),
			"currSlot":   match.GetCurrSlot(),
			"winner":     "",
		}

		if match.IsGameOver() {
			response = map[string]interface{}{
				"message":    "Game Over",
				"board":      match.GetBoard(),
				"currPlayer": match.GetCurrPlayer(),
				"winner":     match.GetWinner(),
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
