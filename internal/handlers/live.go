package handlers

import (
	"4connect/internal/game"
	"4connect/internal/models"
	"4connect/internal/store"
	"4connect/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var connectionLookup = map[string][]models.PlayerConnection{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleJoinMessage(conn *websocket.Conn, msg models.Message, match *game.Game) {

	fmt.Printf("some one has joined game %s \n", msg.MatchID)
	joinMessage := map[string]string{"message": "Joined"}
	conn.WriteJSON(joinMessage)
	conns, ok := connectionLookup[msg.MatchID]

	if !ok {
		connectionLookup[msg.MatchID] = []models.PlayerConnection{{Player: msg.Player, Conn: conn, Time: time.Now().UnixMilli()}}
		return
	}

	if len(conns) == 1 {
		fmt.Printf("we have the second player signup")
		conns = append(conns, models.PlayerConnection{
			Player: msg.Player,
			Conn:   conn,
			Time:   time.Now().UnixMilli()})
		connectionLookup[msg.MatchID] = conns
		game.UpdateNames(match, conns[0].Player, conns[1].Player)

		for _, c := range conns {
			startMessage := map[string]interface{}{
				"message":    "Game Started",
				"board":      game.GetBoard(match),
				"currPlayer": game.GetCurrPlayer(match),
				"player1":    match.Player1,
				"player2":    match.Player2,
			}
			c.Conn.WriteJSON(startMessage)
		}
		return
	}
	// reconnect logic
	conns = connectionLookup[msg.MatchID]
	if conns[0].Player == msg.Player {
		conns[0].Conn = conn
		conns[0].Time = time.Now().UnixMilli()
	} else if conns[1].Player == msg.Player {
		conns[1].Conn = conn
		conns[1].Time = time.Now().UnixMilli()
	} else {
		conn.Close()
		return
	}
	startMessage := map[string]interface{}{
		"message":    "Game Started",
		"board":      game.GetBoard(match),
		"currPlayer": game.GetCurrPlayer(match),
		"player1":    match.Player1,
		"player2":    match.Player2,
	}
	conn.WriteJSON(startMessage)
}

func handlePingMessage(conn *websocket.Conn, msg models.Message) {
	data, ok := connectionLookup[msg.MatchID]

	if !ok {
		fmt.Printf("Match id: %s does not exists", msg.MatchID)
		conn.Close()
	}

	if data[0].Player == msg.Player {
		data[0].Time = time.Now().UnixMilli()
	} else {
		data[1].Time = time.Now().UnixMilli()
	}

	if len(data) == 1 {
		return
	}
	// dead connection so close
	if utils.AbsInt64(data[0].Time-data[1].Time) > 15000 {
		gameOverMessage := map[string]interface{}{
			"message": "Game Over",
			"winner":  msg.Player,
		}
		conn.WriteJSON(gameOverMessage)
		for _, c := range data {
			c.Conn.Close()
		}
		return
	}
	// response with a pong
	response := map[string]interface{}{"message": "pong"}
	conn.WriteJSON(response)
}

func handleMoveMessage(conn *websocket.Conn, msg models.Message, match *game.Game) {

	data, ok := connectionLookup[msg.MatchID]

	if !ok {
		fmt.Printf("Match id: %s does not exists", msg.MatchID)
		conn.Close()
	}

	err := game.MakeMove(match, msg.Player, msg.Move)

	if err != nil {
		failedMove := map[string]interface{}{
			"message":    err.Error(),
			"board":      game.GetBoard(match),
			"currPlayer": game.GetCurrPlayer(match),
		}
		conn.WriteJSON(failedMove)
	}

	updateMessage := map[string]interface{}{
		"message":    "Update Game",
		"board":      game.GetBoard(match),
		"currPlayer": game.GetCurrPlayer(match),
	}
	for _, c := range data {
		c.Conn.WriteJSON(updateMessage)
	}

	if game.IsGameOver(match) {
		gameOverMessage := map[string]interface{}{
			"message": "Game Over",
			"winner":  game.GetWinner(match),
		}
		for _, c := range data {
			c.Conn.WriteJSON(gameOverMessage)
		}
	}
}

func handleRematchMessage(conn *websocket.Conn, msg models.Message) {
	data, ok := connectionLookup[msg.MatchID]
	if !ok {
		fmt.Printf("Match id: %s does not exists", msg.MatchID)
		conn.Close()
	}

	if msg.Message == "request" {
		for _, c := range data {
			if c.Player != msg.Player {
				requestMessage := map[string]interface{}{
					"message": "ReMatch Request",
				}
				c.Conn.WriteJSON(requestMessage)
			}
		}
	} else {
		if msg.Message != "accept" {
			for _, c := range data {
				c.Conn.Close()
			}
			return
		}
		datastore := store.GetDataStore()
		datastore[msg.MatchID] = game.NewGame(data[1].Player, data[0].Player)
		match := datastore[msg.MatchID]
		for _, c := range data {
			startMessage := map[string]interface{}{
				"message":    "Game Started",
				"board":      game.GetBoard(match),
				"currPlayer": game.GetCurrPlayer(match),
				"player1":    match.Player1,
				"player2":    match.Player2,
			}
			c.Conn.WriteJSON(startMessage)
		}
	}
}

func wsMessageHandler(conn *websocket.Conn) {
	defer conn.Close()

	for {
		// Read JSON message from WebSocket
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		log.Printf("type: %s, message: %s", msg.Type, msg.Message)
		datastore := store.GetDataStore()
		match, ok := datastore[msg.MatchID]
		if !ok {
			fmt.Println("Match does not exists")
			conn.Close()
			return
		}
		switch msg.Type {
		case models.JoinMessageType:
			handleJoinMessage(conn, msg, match)
		case models.MoveMessageType:
			handleMoveMessage(conn, msg, match)
		case models.PingMessageType:
			handlePingMessage(conn, msg)
		case models.RematchMessageType:
			handleRematchMessage(conn, msg)
		default:
			fmt.Println("Unknown message type:", msg.Type)
		}
	}
}

func LiveWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	wsMessageHandler(conn)
}

func LivePageHandler(w http.ResponseWriter, r *http.Request) {
	matchID := strings.TrimPrefix(r.URL.Path, "/live/")

	datastore := store.GetDataStore()
	_, ok := datastore[matchID]
	if !ok {
		response := map[string]string{
			"message": "match does not exists",
		}
		utils.ReturnJson(w, response, http.StatusBadRequest)
		return
	}
	var data = models.LivePageData{MatchID: matchID}
	utils.RenderTemplate(w, "live.html", data)
}
