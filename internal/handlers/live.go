package handlers

import (
	"4connect/internal/models"
	"4connect/internal/store"
	"4connect/internal/utils"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleLobbyJoinMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {
	matchStore := store.MatchManagerFactory()
	_, err := matchStore.GetMatch(matchId)
	if err != nil {
		errorMessage := map[string]string{"message": "Match does not exists"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	sessionPlayerMap := store.SessionPlayerFactory()
	value, ok := sessionPlayerMap.Load(sessionId)
	if !ok {
		errorMessage := map[string]string{"message": "Session does not exists"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}
	playerId := value.(string)

	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		errorMessage := map[string]string{"message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	message := map[string]string{}
	_, err = session.GetConnection(sessionId)
	if err == nil {
		// Reconnect the old session
		_, err = session.UpdateConnection(sessionId, playerId, conn)
		if err != nil {
			errorMessage := map[string]string{"message": err.Error()}
			fmt.Println(errorMessage)
			conn.WriteJSON(errorMessage)
			conn.Close()
		}
		message = map[string]string{"message": "Reconnected to old session"}
	} else {
		fmt.Println("Adding new Player to the Session", sessionId, playerId)
		_, err = session.AddConnection(sessionId, playerId, conn)
		if err != nil {
			errorMessage := map[string]string{"message": err.Error()}
			fmt.Println(errorMessage)
			conn.WriteJSON(errorMessage)
			conn.Close()
			return
		}
		message = map[string]string{"message": "joined session"}
	}
	fmt.Println(message)
	conn.WriteJSON(message)
}

func handleGameJoinMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {

	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		errorMessage := map[string]string{"message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	player, err := session.GetConnection(sessionId)
	if err != nil {
		errorMessage := map[string]string{"message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	if player.Conn != conn {
		errorMessage := map[string]string{"message": "Session is not associated with this connection"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}
	fmt.Println("Adding Player to Game", sessionId, msg.Name)
	player, err = session.AddPlayerToGame(sessionId, msg.Name)
	if err != nil {
		errorMessage := map[string]string{"message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	message := map[string]string{"message": "joined game", "player": player.Name, "type": player.Type}
	fmt.Println("Sending to message queue: ", message)
	session.AddToMessageQueue(message)
	conn.WriteJSON(message)
	session.BroadcastGameState()
}

func handlePingMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {
	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	player, err := session.GetConnection(sessionId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	player.Time = time.Now().UnixMilli()

	if session.State == models.SessionStateLive && player.Type == models.ActivePlayerType {
		player1, err := session.GetPlayer(session.Game.Player1)
		if err != nil {
			fmt.Println(err)
			message := map[string]interface{}{
				"message": "Game Over",
				"winner":  session.Game.GetWinner(),
			}
			session.BroadcastMessage(message)
			return
		}
		player2, err := session.GetPlayer(session.Game.Player2)
		if err != nil {
			fmt.Println(err)
			message := map[string]interface{}{
				"message": "Game Over",
				"winner":  session.Game.GetWinner(),
			}
			session.BroadcastMessage(message)
			return
		}
		if utils.AbsInt64(player1.Time-player2.Time) > 15000 {
			session.State = models.SessionStateClosed
			message := map[string]interface{}{
				"message": "Game Over",
				"winner":  session.Game.GetWinner(),
			}
			session.BroadcastMessage(message)
			return
		}
		response := map[string]interface{}{"message": "pong"}
		conn.WriteJSON(response)
	}
}

func handleMoveMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {

	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	player, err := session.GetConnection(sessionId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	if player.Type != models.ActivePlayerType {
		errorMessage := map[string]string{"message": "You are not the active player"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		return
	}

	err = session.Game.MakeMove(player.ID, msg.Move)

	if err != nil {
		failedMove := map[string]interface{}{
			"message": "Failed to make move",
		}
		conn.WriteJSON(failedMove)
		return
	}
	session.BroadcastGameState()
	if session.Game.IsGameOver() {
		session.State = models.SessionStateRematch
		message := map[string]interface{}{
			"message": "Game Over",
			"winner":  session.GetGameWinner(),
		}
		session.BroadcastMessage(message)
	}
}

func handleChatMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {
	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}
	player, err := session.GetConnection(sessionId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	message := map[string]interface{}{
		"type":    models.ChatMessageType,
		"message": msg.Message,
		"player":  player.Name,
	}
	session.BroadcastMessage(message)
}

func handleRematchMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {
	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	player, err := session.GetConnection(sessionId)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}
	if player.Type != models.ActivePlayerType {
		errorMessage := map[string]string{"message": "You are not the active player"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		return
	}

	// if msg.Message == "request" {
	// requestMessage := map[string]interface{}{
	// 	for _, c := range data {
	// 		c.Conn.Close()
	// 	}
	// 	return
	// }
	// datastore := store.MatchManagerFactory()
	// gameState, err := datastore.ResetGame(msg.MatchID, true)
	// if err != nil {
	// 	for _, c := range data {
	// 		c.Conn.Close()
	// 	}
	// 	return
	// }
	// for _, c := range data {
	// 	startMessage := map[string]interface{}{
	// 		"message":    "Game Started",
	// 		"board":      gameState.Board,
	// 		"currPlayer": gameState.CurrPlayer,
	// 		"player1":    gameState.PlayerAName,
	// 		"player2":    gameState.PlayerBName,
	// 	}
	// 	c.Conn.WriteJSON(startMessage)
	// }
	// }
}

func wsMessageHandler(conn *websocket.Conn, matchId string, sessionId string) {
	defer conn.Close()

	for {
		// Read JSON message from WebSocket
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		matchStore := store.MatchManagerFactory()
		_, err = matchStore.GetMatch(matchId)
		if err != nil {
			fmt.Println("Match does not exists")
			conn.Close()
			return
		}

		switch msg.Type {
		case models.JoinSessionMessageType:
			handleLobbyJoinMessage(conn, matchId, sessionId, msg)
		case models.JoinMatchMessageType:
			handleGameJoinMessage(conn, matchId, sessionId, msg)
		case models.MoveMessageType:
			handleMoveMessage(conn, matchId, sessionId, msg)
		case models.PingMessageType:
			handlePingMessage(conn, matchId, sessionId, msg)
		case models.ChatMessageType:
			handleChatMessage(conn, matchId, sessionId, msg)
		case models.RematchMessageType:
			handleRematchMessage(conn, matchId, sessionId, msg)
		default:
			fmt.Println("Unknown message type:", msg.Type)
			conn.Close()
			return
		}
	}
}

func LiveWebSocketHandler(w http.ResponseWriter, r *http.Request) {

	matchId := r.URL.Query().Get("matchId")
	sessionId := r.URL.Query().Get("sessionId")

	if matchId == "" || sessionId == "" {
		fmt.Println("Match id does not exits")
		response := map[string]interface{}{
			"message": "Incomplete Request",
		}
		utils.ReturnJson(w, response, http.StatusBadRequest)
		return
	}

	matchStore := store.MatchManagerFactory()
	_, err := matchStore.GetMatch(matchId)

	if err != nil {
		fmt.Println("Match id does not exits")
		response := map[string]interface{}{
			"message": "Match not found",
		}
		utils.ReturnJson(w, response, http.StatusNotFound)
		return
	}

	sessionManager := store.SessionFactory()
	_, err = sessionManager.GetSession(matchId)
	if err != nil {
		fmt.Println("Session does not exist for match", matchId)
		response := map[string]interface{}{
			"message": "Match not found",
		}
		utils.ReturnJson(w, response, http.StatusNotFound)
		return
	}

	sessionPlayerMap := store.SessionPlayerFactory()
	_, ok := sessionPlayerMap.Load(sessionId)
	if !ok {
		response := map[string]interface{}{
			"message": "Session does not exists",
		}
		utils.ReturnJson(w, response, http.StatusMethodNotAllowed)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	go wsMessageHandler(conn, matchId, sessionId)
}
