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
	Player1    string
	Player2    string
	Board      []int
	Winner     string
	CurrPlayer string
	CurrSlot   string
	State      string
}

type MatchManager struct {
	Matches map[string]*models.MatchSession
	mu      sync.Mutex
}

func MakeMatchManager() *MatchManager {
	return &MatchManager{Matches: make(map[string]*models.MatchSession), mu: sync.Mutex{}}
}

func (manager *MatchManager) CreateGame(playerAName string, playerBName string, level string, gameType string) string {

	manager.mu.Lock()
	defer manager.mu.Unlock()
	var matchId string = utils.GenerateMatchId(manager.Matches)
	var match *game.Game = game.NewGame(playerAName, playerBName)
	var session = &models.MatchSession{Game: match, Level: level,
		GameType: gameType, State: "waiting",
		Player1Status: false,
		Player2Status: false,
	}
	manager.Matches[matchId] = session

	return matchId
}

func (manager *MatchManager) GetSessionState(matchId string) string {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	session, ok := manager.Matches[matchId]
	if !ok {
		return "ended"
	}
	return session.State
}

func (manager *MatchManager) SignUpPlayer(matchId string, player string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	session, ok := manager.Matches[matchId]
	if !ok {
		return errors.New("Match does not exists")
	}

	if session.State != "waiting" {
		return errors.New("Match cannot sign up any more player")
	}

	if !session.Player1Status {
		session.Game.Player1 = player
		session.Player1Status = true

		if session.GameType == "bot" {
			session.State = "playing"
		}
		return nil
	}

	if session.GameType == "live" && !session.Player2Status {
		session.Game.Player2 = player
		session.Player2Status = true

		session.State = "playing"
		return nil
	}
	return errors.New("Failed to sign up player")
}

func (manager *MatchManager) MakeMove(matchId string, playerId string, move int) error {
	match, err := manager.GetMatch(matchId)
	if err != nil {
		return err
	}

	if match.Player1 == playerId {
		return game.MakeMove(match, "RED", move)
	}
	return game.MakeMove(match, "YELLOW", move)
}

func (manager *MatchManager) GetMatchSession(matchID string) (*models.MatchSession, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	match, ok := manager.Matches[matchID]
	if !ok {
		return &models.MatchSession{}, errors.New("Match does not exist")
	}
	return match, nil
}

func (manager *MatchManager) GetMatchState(matchID string) (MatchState, error) {

	match, err := manager.GetMatchSession(matchID)
	if err != nil {
		return MatchState{}, err
	}
	board := match.Game.GetBoard()
	winner := match.Game.GetWinner()

	return MatchState{
		Player1:    match.Game.Player1,
		Player2:    match.Game.Player2,
		Board:      board,
		Winner:     winner,
		State:      match.State,
		CurrPlayer: match.Game.GetCurrPlayer(),
		CurrSlot:   match.Game.GetCurrSlot(),
	}, nil

}

func (manager *MatchManager) GetMatch(id string) (*game.Game, error) {
	match, ok := manager.Matches[id]
	if !ok {
		return &game.Game{}, errors.New("Match does not exist")
	}

	return match.Game, nil
}

func (manager *MatchManager) ResetGame(id string, flip bool) (MatchState, error) {
	match, ok := manager.Matches[id]
	if !ok {
		return MatchState{}, errors.New("Match does not exist")
	}

	if flip {
		match.Game = game.NewGame(match.Game.Player2, match.Game.Player1)
	} else {
		match.Game = game.NewGame(match.Game.Player1, match.Game.Player2)
	}

	return manager.GetMatchState(id)
}
