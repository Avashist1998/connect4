package models

import (
	"4connect/internal/game"
	"fmt"
	"math/rand"

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
	SessionStateMessageType = "session_state"
	GameStateMessageType    = "game_state"
	GameOverMessageType     = "game_over"
	ErrorMessageType        = "error"
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

type RematchSession struct {
	Player1Accepted bool
	Player2Accepted bool
	Time            int64
}

type Session struct {
	Game        *game.Game
	State       string             // Session States: init, live, ended
	connections map[string]*Player // Mapping to session id to player
	players     map[string]*Player // Mapping to player id to player
	messages    chan interface{}
	Rematch     *RematchSession // Rematch only exists when in rematch state
	matchCount  int             // Count of matches played in the session
	player1Wins int             // Count of wins for player 1
	player2Wins int             // Count of wins for player 2
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
		Rematch:     nil,
		matchCount:  0,
		player1Wins: 0,
		player2Wins: 0,
	}
}

func (s *Session) StartNewGame(randomize bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// randomize the player1 and player2
	var newGame *game.Game
	var err error
	if randomize {
		if rand.Intn(2) == 0 {
			newGame, err = game.NewGame(s.Game.Player1, s.Game.Player2)
		} else {
			newGame, err = game.NewGame(s.Game.Player2, s.Game.Player1)
		}
	} else {
		newGame, err = game.NewGame(s.Game.Player1, s.Game.Player2)
	}
	if err != nil {
		return
	}
	s.Game = newGame
	s.State = SessionStateLive
}

func (s *Session) ProcessRematchRequest(playerId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rematch == nil {
		s.Rematch = &RematchSession{
			Player1Accepted: false,
			Player2Accepted: false,
			Time:            time.Now().UnixMilli(),
		}
	}
	if playerId == s.Game.Player1 {
		s.Rematch.Player1Accepted = true
	} else {
		s.Rematch.Player2Accepted = true
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

func (s *Session) AddPlayerToGame(sessionId string, name string) (*Player, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.connections[sessionId]
	if !ok {
		return nil, false, errors.New("Player does not exist")
	}
	// Set player name first
	player.Name = name

	gameStarted := false
	if s.Game == nil {
		game, err := game.NewGame(player.ID, "")
		if err != nil {
			return nil, false, errors.New("Failed to create game")
		}
		s.Game = game
		player.Type = ActivePlayerType

		fmt.Println("Game Player 1: ", player.Name, "ID:", player.ID)
	} else if s.Game.Player2 == "" && s.Game.Player1 != "" && s.Game.Player1 != player.ID {
		s.Game.Player2 = player.ID
		s.State = SessionStateLive
		player.Type = ActivePlayerType
		gameStarted = true
		fmt.Println("Game Player 2: ", player.Name)
		fmt.Println("Game started", s.Game.Player1, s.Game.Player2)
	} else {
		gameStarted = true
		fmt.Println("Player is already in the game or game is already started")
	}
	return player, gameStarted, nil
}

func (s *Session) GetState() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.State
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

func (s *Session) GetPlayerStatus() (string, string) {
	player1Status := "live"
	player2Status := "live"
	player1, err := s.GetPlayer(s.Game.Player1)
	if err != nil {
		player1Status = "disconnected"
	}
	player2, err := s.GetPlayer(s.Game.Player2)
	if err != nil {
		player2Status = "disconnected"
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if player1 != nil && player1.Time-time.Now().UnixMilli() > 15000 {
		player1Status = "disconnected"
	}
	if player2 != nil && player2.Time-time.Now().UnixMilli() > 15000 {
		player2Status = "disconnected"
	}
	return player1Status, player2Status
}

func (s *Session) GetSessionState() (map[string]interface{}, error) {
	player1Status, player2Status := s.GetPlayerStatus()
	return map[string]interface{}{
		"type":             SessionStateMessageType,
		"state":            s.State,
		"match_count":      s.matchCount,
		"player1_wins":     s.player1Wins,
		"player2_wins":     s.player2Wins,
		"player1Status":    player1Status,
		"player2Status":    player2Status,
		"connection_count": s.GetConnectionCount(),
	}, nil
}

func (s *Session) BroadcastSessionState() {
	message, err := s.GetSessionState()
	fmt.Println("We are going to broadcast the session state: ", message)
	if err != nil {
		return
	}
	s.BroadcastMessage(message)
}

func (s *Session) BroadcastGameState() {
	// Locking
	s.mu.Lock()
	defer s.mu.Unlock()

	// Safety check: game must exist and have both players
	if s.Game == nil {
		fmt.Println("BroadcastGameState: Game is nil")
		return
	}
	if s.Game.Player1 == "" || s.Game.Player2 == "" {
		fmt.Println("BroadcastGameState: Game does not have both players yet. Player1:", s.Game.Player1, "Player2:", s.Game.Player2)
		return
	}
	fmt.Println("Broadcasting Game State")
	currentPlayer := ""
	player1, ok := s.players[s.Game.Player1]
	if !ok {
		fmt.Println("Player 1 does not exist")
		return
	}
	fmt.Println("Broadcasting Game State Player 1")
	fmt.Println("Player 1: ", player1.Name)
	player2, ok := s.players[s.Game.Player2]
	if !ok {
		fmt.Println("Player 2 does not exist")
		return
	}
	fmt.Println("Broadcasting Game State Player 2")
	if s.Game.GetCurrPlayer() == s.Game.Player1 {
		currentPlayer = "player1"
	} else {
		currentPlayer = "player2"
	}
	winner := ""
	if s.Game.GetWinner() == s.Game.Player1 {
		winner = "player1"
	} else if s.Game.GetWinner() == s.Game.Player2 {
		winner = "player2"
	}
	message := map[string]interface{}{
		"type":             GameStateMessageType,
		"board":            s.Game.GetBoard(),
		"currPlayer":       currentPlayer,
		"player1":          player1.Name,
		"player2":          player2.Name,
		"player1Color":     "RED",
		"player2Color":     "YELLOW",
		"winner":           winner,
		"state":            s.State,
		"connection_count": len(s.connections),
	}
	s.AddToMessageQueue(message)
	// Collect connections and player info while holding the lock
	connections := make([]struct {
		Conn     *websocket.Conn
		PlayerID string
	}, 0, len(s.connections))
	for _, player := range s.connections {
		if player.Conn != nil {
			connections = append(connections, struct {
				Conn     *websocket.Conn
				PlayerID string
			}{player.Conn, player.ID})
		}
	}

	for _, player := range connections {
		switch {
		case player.PlayerID == s.Game.Player1:
			message["you"] = "player1"
		case player.PlayerID == s.Game.Player2:
			message["you"] = "player2"
		default:
			message["you"] = "none"
		}
		// Safely write with error handling
		if err := player.Conn.WriteJSON(message); err != nil {
			// Connection is likely closed, log and continue
			fmt.Printf("Error writing to connection for player %s: %v\n", player.PlayerID, err)
			continue
		}
	}

}

func (s *Session) BroadcastMessage(message interface{}) {
	fmt.Println("Broadcasting Message", message)
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println("Lock acquired")
	connections := make([]struct {
		Conn     *websocket.Conn
		PlayerId string
	}, 0, len(s.players))
	for _, player := range s.players {
		if player.Conn != nil {
			connections = append(connections, struct {
				Conn     *websocket.Conn
				PlayerId string
			}{player.Conn, player.ID})
		}
	}

	for _, connection := range connections {
		if err := connection.Conn.WriteJSON(message); err != nil {
			fmt.Println("Failed to write message to connection: to Player id ", connection.PlayerId, " with error: ", err)
			continue
		}
	}
	fmt.Println("Message broadcasted")
}

func (s *Session) SendMessageToPlayer(playerId string, message interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.players[playerId]
	if !ok {
		return errors.New("Player does not exist")
	}
	player.Conn.WriteJSON(message)
	return nil
}

func (s *Session) SendMessageToSession(sessionId string, message interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	player, ok := s.connections[sessionId]
	if !ok {
		return errors.New("Player does not exist")
	}
	player.Conn.WriteJSON(message)
	return nil
}

func (s *Session) ProcessGameOver() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State = SessionStateRematch
	s.matchCount++
	// Need to fix this
	if s.Game.GetWinner() == s.Game.Player1 {
		s.player1Wins++
	} else {
		s.player2Wins++
	}
}

func (s *Session) GetGameOverMessage() map[string]interface{} {

	if s.Game.GetWinner() == "" {
		return map[string]interface{}{
			"type":    GameOverMessageType,
			"message": "Game Over",
			"winner":  "draw",
		}
	} else if s.Game.GetWinner() == s.Game.Player1 {
		return map[string]interface{}{
			"type":    GameOverMessageType,
			"message": "Game Over",
			"winner":  "player1",
		}
	} else {
		return map[string]interface{}{
			"type":    GameOverMessageType,
			"message": "Game Over",
			"winner":  "player2",
		}
	}
}

func (s *Session) BroadcastGameOver() {
	message := s.GetGameOverMessage()
	s.BroadcastMessage(message)
}
