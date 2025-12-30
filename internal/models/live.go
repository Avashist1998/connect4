package models

import (
	"4connect/internal/game"
	"fmt"

	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	JoinSessionMessageType  = "join"
	JoinMatchMessageType    = "ready"
	MoveMessageType         = "move"
	PingMessageType         = "ping"
	RematchMessageType      = "rematch"
	GetGameStateMessageType = "get_game_state"
	ChatMessageType         = "chat"
)

type GameState struct {
	Board         []int
	CurrentPlayer string
	Player1       string
	Player2       string
	Player1Color  string
	Player2Color  string
	Winner        string
	State         string
}

type Message struct {
	Type    string `json:"type"`    // Type of message (join, move, ping)
	Move    int    `json:"move"`    // Move (for MoveMessage)
	Name    string `json:"name"`    // Name of the player
	Message string `json:"message"` // Message content (optional)
}

type LivePageData struct {
	MatchID string
}

const (
	ActivePlayerType  = "active"
	PassivePlayerType = "passive"
)

type Player struct {
	ID   string
	Name string
	Conn *websocket.Conn
	Time int64
	Type string // Player Connection Types: active, passive
}

const (
	SessionStateInit    = "init"
	SessionStateLive    = "live"
	SessionStateRematch = "rematch"
	SessionStateClosed  = "closed"
)

type Session struct {
	Game        *game.Game
	State       string             // Session States: init, live, ended
	connections map[string]*Player // Mapping to session id to player
	players     map[string]*Player // Mapping to player id to player
	messages    chan interface{}
	mu          sync.Mutex
}

func MakeSession() *Session {
	return &Session{
		State:       SessionStateInit,
		connections: make(map[string]*Player),
		players:     make(map[string]*Player),
		messages:    make(chan interface{}, 100),
		mu:          sync.Mutex{},
		Game:        nil,
	}
}

func (s *Session) AddToMessageQueue(message interface{}) {
	s.messages <- message
	fmt.Println("Added message to queue: ", message)
}

func (s *Session) GetConnectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.connections)
}

func (s *Session) GetSessionId(conn *websocket.Conn) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sessionId, player := range s.connections {
		if player.Conn == conn {
			return sessionId, nil
		}
	}
	return "", errors.New("Connection not found in session")
}

func (s *Session) AddConnection(sessionId string, playerId string, conn *websocket.Conn) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.connections[sessionId]; ok {
		return nil, errors.New("Player already exists")
	}

	player := &Player{
		ID:   playerId,
		Conn: conn,
		Time: time.Now().UnixMilli(),
		Type: ActivePlayerType,
	}
	s.connections[sessionId] = player
	s.players[playerId] = player
	return player, nil
}

func (s *Session) UpdateConnection(sessionId string, playerId string, conn *websocket.Conn) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.connections[sessionId]
	if !ok {
		return nil, errors.New("Player does not exist")
	}
	_, ok = s.players[playerId]
	if !ok {
		return nil, errors.New("Player does not exist")
	}
	if player.ID != playerId {
		return nil, errors.New("Player ID does not match")
	}
	if player.Conn != nil {
		player.Conn.Close()
	}
	player.Conn = conn
	player.Time = time.Now().UnixMilli()
	return player, nil
}

func (s *Session) GetPlayer(playerId string) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player, ok := s.players[playerId]
	if !ok {
		return nil, errors.New("Player does not exist")
	}
	return player, nil
}

func (s *Session) GetConnection(sessionId string) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.connections[sessionId]
	if !ok {
		return nil, errors.New("Player does not exist")
	}
	return player, nil
}

func (s *Session) AddPlayerToGame(sessionId string, name string) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.connections[sessionId]
	if !ok {
		return nil, errors.New("Player does not exist")
	}

	if s.Game == nil {
		game, err := game.NewGame(player.ID, "")
		if err != nil {
			return nil, errors.New("Failed to create game")
		}
		s.Game = game
		player.Type = ActivePlayerType
		player.Name = name
		fmt.Println("Game Player 1: ", player.Name)
	} else if s.Game.Player2 == "" {
		game, err := game.NewGame(s.Game.Player1, player.ID)
		if err != nil {
			return nil, errors.New("Failed to create game")
		}
		s.Game = game
		player.Type = ActivePlayerType
		player.Name = name
		s.State = SessionStateLive
		fmt.Println("Game Player 2: ", player.Name)
		fmt.Println("Game started", s.Game.Player1, s.Game.Player2)
	} else {
		fmt.Println(s.Game.Player1, s.Game.Player2)

	}
	return player, nil
}

func (s *Session) GetGameWinner() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Game.GetWinner() == s.Game.Player1 {
		return "player1"
	} else if s.Game.GetWinner() == s.Game.Player2 {
		return "player2"
	}
	return ""
}

func (s *Session) BroadcastGameState() {
	// Locking

	fmt.Println("Broadcasting Game State")
	currentPlayer := ""
	player1, err := s.GetPlayer(s.Game.Player1)
	if err != nil {
		return
	}
	fmt.Println("Broadcasting Game State Player 1")
	fmt.Println("Player 1: ", player1.Name)
	fmt.Println("Player 2 ID: ", s.Game.Player2)
	player2, err := s.GetPlayer(s.Game.Player2)
	if err != nil {
		return
	}
	fmt.Println("Broadcasting Game State Player 2")
	if s.Game.GetCurrPlayer() == s.Game.Player1 {
		currentPlayer = "player1"
	} else {
		currentPlayer = "player2"
	}

	winner := s.GetGameWinner()
	message := map[string]interface{}{
		"message":          "Game State",
		"board":            s.Game.GetBoard(),
		"currPlayer":       currentPlayer,
		"player1":          player1.Name,
		"player2":          player2.Name,
		"player1Color":     "RED",
		"player2Color":     "YELLOW",
		"winner":           winner,
		"state":            s.State,
		"connection_count": s.GetConnectionCount(),
	}
	s.AddToMessageQueue(message)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, player := range s.connections {
		switch {
		case player.ID == s.Game.Player1:
			message["you"] = "player1"
			player.Conn.WriteJSON(message)
		case player.ID == s.Game.Player2:
			message["you"] = "player2"
			player.Conn.WriteJSON(message)
		default:
			message["you"] = "none"
			player.Conn.WriteJSON(message)
		}
	}
}

func (s *Session) BroadcastMessage(message interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, player := range s.connections {
		player.Conn.WriteJSON(message)
	}
}
