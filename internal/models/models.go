package models

import "html/template"

type MatchData struct {
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
