package store

import (
	"4connect/internal/services"
)

var gameLobby *services.Lobby
var matchManager *services.MatchManager

func MatchManagerFactory() *services.MatchManager {
	if matchManager == nil {
		matchManager = services.MakeMatchManager()
	}
	return matchManager
}

func LobbyFactory() *services.Lobby {
	if gameLobby == nil {
		gameLobby = services.MakeLobby()
	}
	return gameLobby
}
