package models

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type LivePlayer struct {
	ID          string
	Name        string
	Kind        string
	Conn        *websocket.Conn
	MessageChan chan map[string]interface{}
}

func (p *LivePlayer) ListenToMessages() {

	for msg := range p.MessageChan {
		// Write the message to the WebSocket
		// fmt.Printf("Message recieved :%s", msg)
		err := p.Conn.WriteJSON(msg) // Assuming text message
		if err != nil {
			fmt.Printf("Error sending message to player %s: %v\n", p.ID, err)
			return
		}
	}
}
