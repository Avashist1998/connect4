package handlers

import (
	"4connect/internal/store"
	"4connect/internal/utils"
	"fmt"
	"net/http"
)

func HandleSessionGet(w http.ResponseWriter, r *http.Request) {
	sessionPlayerMap := store.SessionPlayerFactory()
	sessionId := r.PathValue("sessionId")
	_, ok := sessionPlayerMap.Load(sessionId)
	if !ok {
		fmt.Println("Session not found", sessionId)

		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	utils.ReturnJson(w, map[string]interface{}{"session_id": sessionId}, http.StatusOK)
}

func HandleSessionPost(w http.ResponseWriter, r *http.Request) {
	sessionPlayerMap := store.SessionPlayerFactory()
	sessionId := utils.GenerateId(10)
	playerId := utils.GenerateId(10)
	sessionPlayerMap.Store(sessionId, playerId)
	utils.ReturnJson(w, map[string]interface{}{"session_id": sessionId}, http.StatusOK)
}
