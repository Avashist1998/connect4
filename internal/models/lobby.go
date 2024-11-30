package models

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type LobbyPlayer struct {
	ID       string
	Conn     *websocket.Conn
	LobbyMsg chan string
}

func (p *LobbyPlayer) ListenToMessages() {

	for msg := range p.LobbyMsg {
		// Write the message to the WebSocket
		err := p.Conn.WriteMessage(1, []byte(msg)) // Assuming text message
		if err != nil {
			fmt.Printf("Error sending message to player %s: %v\n", p.ID, err)
			return
		}
	}
}
