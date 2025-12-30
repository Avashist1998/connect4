package models

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type LobbyPlayer struct {
	ID       string
	Conn     *websocket.Conn
	LobbyMsg chan string
	writeMu  sync.Mutex // Protects Conn.WriteMessage calls
}

type LobbyConnection struct {
	SessionId string
	Conn      *websocket.Conn
	Time      int64
	Type      string
}

type LobbyMessage struct {
	Type    string `json:"type"`
	Player  string `json:"player"`  // Player's name
	Message string `json:"message"` // Message content (optional)
}

func (p *LobbyPlayer) ListenToMessages() {
	for msg := range p.LobbyMsg {
		// Write the message to the WebSocket (protected by mutex to prevent concurrent writes)
		p.writeMu.Lock()
		err := p.Conn.WriteMessage(1, []byte(msg)) // Assuming text message
		p.writeMu.Unlock()
		if err != nil {
			fmt.Printf("Error sending message to player %s: %v\n", p.ID, err)
			return
		}
	}
}
