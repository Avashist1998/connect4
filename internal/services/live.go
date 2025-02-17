package services

import (
	"4connect/internal/models"
	"4connect/internal/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type LiveSessionManager struct {
	sessions map[string][]*models.LivePlayer
	mu       sync.Mutex
}

func MakeLiveSessionManager() *LiveSessionManager {
	return &LiveSessionManager{
		sessions: make(map[string][]*models.LivePlayer),
		mu:       sync.Mutex{}}
}

func (manager *LiveSessionManager) AddPlayer(
	matchID string,
	player string,
	conn *websocket.Conn,
	matchManager *MatchManager) string {

	manager.mu.Lock()
	defer manager.mu.Unlock()

	players, ok := manager.sessions[matchID]
	if !ok {
		manager.sessions[matchID] = []*models.LivePlayer{}
	}
	players, _ = manager.sessions[matchID]
	playerId := utils.GenerateId(6)

	kind := "passive"
	err := matchManager.SignUpPlayer(matchID, playerId)
	if err != nil {
		kind = "active"
	}
	livePlayer := &models.LivePlayer{
		ID:          playerId,
		Name:        player,
		Kind:        kind,
		Conn:        conn,
		MessageChan: make(chan map[string]interface{})}

	manager.sessions[matchID] = append(players, livePlayer)
	players, ok = manager.sessions[matchID]

	matchState, err := matchManager.GetMatchState(matchID)
	if playerId == matchState.Player1 || playerId == matchState.Player2 {
		go manager.listenToPlayer(matchID, livePlayer, matchManager)
	}
	go livePlayer.ListenToMessages()

	joinMessage := map[string]interface{}{
		"type":       "Join Player",
		"message":    "Player Joined",
		"playerId":   livePlayer.ID,
		"playerName": livePlayer.Name,
		"joinTime":   time.Now().UTC().UnixMilli()}

	log.Printf("%s\n", joinMessage)
	livePlayer.MessageChan <- joinMessage

	log.Printf("The current state of the game %s, and error %s", matchState, err)
	if err == nil && matchState.State == "playing" {
		player1Name := ""
		player2Name := ""
		for _, p := range players {
			if p.ID == matchState.Player1 {
				player1Name = p.Name
			}
			if p.ID == matchState.Player2 {
				player2Name = p.Name
			}
		}
		res := map[string]interface{}{
			"message":    "Update Game",
			"player1":    player1Name,
			"player2":    player2Name,
			"board":      matchState.Board,
			"currPlayer": matchState.CurrPlayer,
			"currSlot":   matchState.CurrSlot,
		}
		for _, p := range players {
			p.MessageChan <- res
		}
	}
	return playerId
}

func (manager *LiveSessionManager) RemovePlayer(matchID string, playerID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	players, ok := manager.sessions[matchID]

	if !ok {
		return
	}

	for i, p := range players {
		if p.ID == playerID {
			p.Conn.Close()
			manager.sessions[matchID] = append(players[:i], players[i+1:]...)
			return
		}
	}
}

func (manager *LiveSessionManager) BroadcastMessage(matchId string, playerID string, message map[string]interface{}) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	players, ok := manager.sessions[matchId]

	if !ok {
		return
	}

	for _, p := range players {
		if p.ID != playerID {
			p.MessageChan <- message
		}
	}
}

func (manager *LiveSessionManager) listenToPlayer(matchId string, player *models.LivePlayer, matchManger *MatchManager) {
	defer manager.RemovePlayer(matchId, player.ID)
	for {
		var msg models.Message
		err := player.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
		switch msg.Type {
		case models.MoveMessageType:
			fmt.Printf("Move message %s \n", msg)
			err = matchManger.MakeMove(matchId, player.ID, msg.Move)
			if err == nil {
				matchState, err := matchManger.GetMatchState(matchId)
				if err == nil {
					player1Name := ""
					player2Name := ""
					players, _ := manager.sessions[matchId]
					for _, p := range players {
						if p.ID == matchState.Player1 {
							player1Name = p.Name
						}
						if p.ID == matchState.Player2 {
							player2Name = p.Name
						}
					}
					res := map[string]interface{}{
						"message":    "Update Game",
						"player1":    player1Name,
						"player2":    player2Name,
						"board":      matchState.Board,
						"currPlayer": matchState.CurrPlayer,
						"currSlot":   matchState.CurrSlot,
					}
					manager.BroadcastMessage(matchId, "", res)
				} else {
					res := map[string]interface{}{
						"message":    err.Error(),
						"board":      matchState.Board,
						"currPlayer": matchState.CurrPlayer,
						"currSlot":   matchState.CurrSlot,
					}
					player.MessageChan <- res
				}
			}
		case models.PingMessageType:
			response := map[string]interface{}{"message": "pong"}
			player.Conn.WriteJSON(response)
		case models.RematchMessageType:
			// send message to the other active player
		}

		matchState, err := matchManger.GetMatchState(matchId)
		if err == nil && matchState.Winner != "" {
			gameOverMessage := map[string]interface{}{
				"message": "Game Over",
				"winner":  matchState.Winner,
			}
			manager.BroadcastMessage(matchId, "", gameOverMessage)
		}
	}
}
