package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"4connect/internal/models"

	"github.com/gorilla/websocket"
)

type Lobby struct {
	Players      map[string]*models.LobbyPlayer
	WaitingQueue []*models.LobbyPlayer
	mu           sync.Mutex
}

func MakeLobby() *Lobby {
	return &Lobby{Players: make(map[string]*models.LobbyPlayer), WaitingQueue: []*models.LobbyPlayer{}, mu: sync.Mutex{}}
}

func (l *Lobby) AddPlayer(id string, conn *websocket.Conn, manager *MatchManager) {
	l.mu.Lock()
	defer l.mu.Unlock()

	player := &models.LobbyPlayer{
		ID:       id,
		Conn:     conn,
		LobbyMsg: make(chan string),
	}

	fmt.Printf("we are adding the player %s\n", player.ID)
	l.Players[id] = player
	l.WaitingQueue = append(l.WaitingQueue, player)

	go l.listenToPlayer(player)
	go player.ListenToMessages()
	joinMessage, err := json.Marshal(map[string]interface{}{
		"type":     "Join Player",
		"message":  "Player Joined",
		"playerId": player.ID,
		"joinTime": time.Now().UTC().UnixMilli()})

	if err != nil {
		joinMessage = []byte{}
	}
	for _, p := range l.Players {
		p.LobbyMsg <- string(joinMessage)
	}
	l.createMatch(manager)
}

func (l *Lobby) RemovePlayer(playerID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	player, exists := l.Players[playerID]
	if !exists {
		return
	}

	leaveMessage, err := json.Marshal(map[string]interface{}{
		"type":     "Leave Player",
		"message":  "Player left the lobby",
		"playerId": player.ID,
	})

	if err == nil {
		for id, player := range l.Players {
			if id != playerID {
				player.LobbyMsg <- string(leaveMessage)
			}
		}
	}

	close(player.LobbyMsg)
	player.Conn.Close()
	delete(l.Players, playerID)
	fmt.Printf("Player %s disconnected\n", playerID)

	for i, player := range l.WaitingQueue {
		if player.ID == playerID {
			l.WaitingQueue = append(l.WaitingQueue[:i], l.WaitingQueue[i+1:]...)
			break
		}
	}
	fmt.Printf("Player %s removed from the lobby\n", playerID)
}

func (l *Lobby) BroadcastMessage(senderID string, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for id, player := range l.Players {
		if id != senderID {
			player.LobbyMsg <- message
		}
	}
}

func (l *Lobby) listenToPlayer(player *models.LobbyPlayer) {
	defer l.RemovePlayer(player.ID)

	for {
		_, msg, err := player.Conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message from player %s: %v\n", player.ID, err)
			return
		}

		fmt.Printf("Received message from player %s: %s\n", player.ID, string(msg))
		// Broadcast the message to other players
		l.BroadcastMessage(player.ID, string(msg))
	}
}

func (l *Lobby) createMatch(manger *MatchManager) {
	if len(l.WaitingQueue) < 2 {
		return
	}
	player1 := l.WaitingQueue[0]
	player2 := l.WaitingQueue[1]
	matchID := manger.CreateGame(player1.ID, player2.ID, "", "live")
	player1.Conn.WriteJSON(map[string]interface{}{"type": "Match Info", "matchID": matchID})
	player2.Conn.WriteJSON(map[string]interface{}{"type": "Match Info", "matchID": matchID})

}
