package models

import (
	"4connect/internal/game"
	"html/template"

	"github.com/gorilla/websocket"
)

type MatchRequestData struct {
	GameType    string
	Player1     string
	Player2     string
	StartPlayer string
}

type MoveData struct {
	Player string
	Move   int
}

type MatchPageData struct {
	Player1     string
	Player2     string
	CurrPlayer  string
	BoardHTML   template.HTML
	NewGameHTML template.HTML
}

type PlayerConnection struct {
	Player string
	Conn   *websocket.Conn
	Time   int64
}

const (
	JoinMessageType    = "join"
	MoveMessageType    = "move"
	PingMessageType    = "ping"
	RematchMessageType = "rematch"
)

type LivePageData struct {
	MatchID string
}

type Message struct {
	MatchID string
	Type    string `json:"type"`    // Type of message (join, move, ping)
	Player  string `json:"player"`  // Player's name
	Move    int    `json:"move"`    // Move (for MoveMessage)
	Message string `json:"message"` // Message content (optional)
}

type Match struct {
	Game     *game.Game
	GameType string
}
