package store

import (
	"4connect/internal/services"
	"sync"
)

var gameLobby *services.Lobby
var matchManager *services.MatchManager
var sessionPlayerMap *sync.Map
var sessionManager *services.SessionManager

func MatchManagerFactory() *services.MatchManager {
	if matchManager == nil {
		matchManager = services.MakeMatchManager()
	}
	return matchManager
}

func SessionPlayerFactory() *sync.Map {
	if sessionPlayerMap == nil {
		sessionPlayerMap = &sync.Map{}
	}
	return sessionPlayerMap
}

func SessionFactory() *services.SessionManager {
	if sessionManager == nil {
		sessionManager = services.MakeSessionManager()
	}
	return sessionManager
}

func LobbyFactory() *services.Lobby {
	if gameLobby == nil {
		gameLobby = services.MakeLobby()
	}
	return gameLobby
}
