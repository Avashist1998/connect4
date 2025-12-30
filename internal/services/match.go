package services

import (
	"4connect/internal/game"
	"4connect/internal/models"
	"4connect/internal/utils"
	"errors"
	"sync"
)

// type MatchSession struct {
// 	PlayerA *models.Player
// 	PlayerB *models.Player
// 	Match   *game.Game
// 	State   string
// }

type MatchState struct {
	PlayerAId   string
	PlayerAName string
	PlayerBId   string
	PlayerBName string
	Board       []int
	Winner      string
	State       string
	CurrPlayer  string
}

type MatchManager struct {
	Matches map[string]*models.MatchSession
	mu      sync.Mutex
}

func MakeMatchManager() *MatchManager {
	return &MatchManager{Matches: make(map[string]*models.MatchSession), mu: sync.Mutex{}}
}

func (manager *MatchManager) CreateGame(playerAName string, playerBName string, level string, gameType string) string {
	var playerBId string
	manager.mu.Lock()
	defer manager.mu.Unlock()
	var matchId string = utils.GenerateMatchId(manager.Matches)
	var playerAId = utils.GenerateId(10)
	playerBId = "bot"
	if gameType != "bot" {
		playerBId = utils.GenerateId(10)
	}
	match, err := game.NewGame(playerAId, playerBId)
	if err != nil {
		return ""
	}
	var session = &models.MatchSession{Game: match, Level: level, GameType: gameType, PlayerAId: playerAId, PlayerBId: playerBId, PlayerAName: playerAName, PlayerBName: playerBName}
	manager.Matches[matchId] = session

	return matchId
}

func (manager *MatchManager) GetSessionData(matchID string) (*models.MatchSession, error) {
	match, ok := manager.Matches[matchID]
	if !ok {
		return &models.MatchSession{}, errors.New("Match does not exist")
	}
	return match, nil
}

func (manager *MatchManager) GetMatchState(id string) (MatchState, error) {
	match, ok := manager.Matches[id]
	if !ok {
		return MatchState{}, errors.New("Match does not exist")
	}
	board := match.Game.GetBoard()
	winner := match.Game.GetWinner()

	return MatchState{
		PlayerAId:   match.Game.Player1,
		PlayerAName: match.Game.Player1,
		PlayerBId:   match.Game.Player2,
		PlayerBName: match.Game.Player2,
		Board:       board,
		Winner:      winner,
		CurrPlayer:  match.Game.GetCurrPlayer(),
	}, nil

}

func (manager *MatchManager) GetMatch(id string) (*game.Game, error) {
	match, ok := manager.Matches[id]
	if !ok {
		return &game.Game{}, errors.New("Match does not exist")
	}

	return match.Game, nil
}

func (manager *MatchManager) GetMatchGameType(id string) (string, error) {
	match, ok := manager.Matches[id]
	if !ok {
		return "", errors.New("Match does not exist")
	}
	return match.GameType, nil
}

func (manager *MatchManager) ResetGame(id string, flip bool) (MatchState, error) {
	matchSession, ok := manager.Matches[id]
	if !ok {
		return MatchState{}, errors.New("Match does not exist")
	}

	if flip {
		newGame, err := game.NewGame(matchSession.PlayerBId, matchSession.PlayerAId)
		if err != nil {
			return MatchState{}, err
		}
		matchSession.Game = newGame
	} else {
		newGame, err := game.NewGame(matchSession.PlayerAId, matchSession.PlayerBId)
		if err != nil {
			return MatchState{}, err
		}
		matchSession.Game = newGame
	}

	return manager.GetMatchState(id)
}
