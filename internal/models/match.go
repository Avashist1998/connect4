package models

import (
	"4connect/internal/game"
	"html/template"
)

type MoveData struct {
	Move     int    `json:"move"`
	PlayerId string `json:"playerId"`
}

type MatchRequestData struct {
	GameType    string
	Player1     string
	Player2     string
	StartPlayer string
	Level       string
}

type MatchPageData struct {
	Player1     string
	Player2     string
	CurrPlayer  string
	BoardHTML   template.HTML
	NewGameHTML template.HTML
}

type MatchSession struct {
	Game        *game.Game
	Level       string
	GameType    string
	PlayerAId   string
	PlayerBId   string
	PlayerAName string
	PlayerBName string
	State       string
}
