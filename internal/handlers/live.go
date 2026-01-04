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
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": "Match does not exists"}
		if err := conn.WriteJSON(errorMessage); err != nil {
			conn.WriteJSON(errorMessage)
			conn.Close()
			return
		}
	}

	sessionPlayerMap := store.SessionPlayerFactory()
	value, ok := sessionPlayerMap.Load(sessionId)
	if !ok {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": "Session does not exists"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}
	playerId := value.(string)
	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	message := map[string]string{}
	_, err = session.GetConnection(sessionId)
	fmt.Println("Error: ", err)
	if err == nil {
		// Reconnect the old session
		fmt.Println("Reconnecting to the old session", sessionId, playerId)
		_, err = session.UpdateConnection(sessionId, playerId, conn)
		if err != nil {
			errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
			fmt.Println(errorMessage)
			conn.WriteJSON(errorMessage)
			conn.Close()
		}
		message = map[string]string{"type": models.JoinSessionMessageType, "message": "reconnected"}
	} else {
		fmt.Println("Adding new Player to the Session", sessionId, playerId)
		_, err = session.AddConnection(sessionId, playerId, conn)
		if err != nil {
			errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
			fmt.Println(errorMessage)
			conn.WriteJSON(errorMessage)
			conn.Close()
			return
		}
		message = map[string]string{"type": models.JoinSessionMessageType, "message": "joined session"}
	}
	fmt.Println("Message: ", message)
	fmt.Println(message)
	conn.WriteJSON(message)
}

func handleGameJoinMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {

	sessionStore := store.SessionFactory()
	session, err := sessionStore.GetSession(matchId)
	if err != nil {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	player, err := session.GetConnection(sessionId)
	if err != nil {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	if player.Conn != conn {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": "Session is not associated with this connection"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}
	fmt.Println("Adding Player to Game", sessionId, msg.Name)
	player, gameStarted, err := session.AddPlayerToGame(sessionId, msg.Name)
	if err != nil {
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": err.Error()}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		conn.Close()
		return
	}

	message := map[string]string{"type": models.JoinMatchMessageType, "message": "joined game", "player": player.Name, "player_type": player.Type}

	fmt.Println("Sending to message queue: ", message)
	session.AddToMessageQueue(message)
	err = session.SendMessageToSession(sessionId, message)
	if err != nil {
		fmt.Println("Error sending message to session: ", err)
		conn.Close()
		return
	}

	// Only broadcast game state if this call to AddPlayerToGame actually started the game
	// This prevents race conditions where both handlers call BroadcastGameState
	if gameStarted {
		fmt.Printf("Player: %s Session id: %s is triggering the game state broadcast\n", player.Name, sessionId)
		session.BroadcastGameState()
	}

	fmt.Printf("I have finished handling the game join message for player: %s session id: %s\n", player.Name, sessionId)
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
	if session.State == models.SessionStateLive {
		// Send pong response to client
		response := map[string]interface{}{"message": "pong"}
		err = session.SendMessageToSession(sessionId, response)
		if err != nil {
			fmt.Println("Error sending message to session: ", err)
			conn.Close()
			return
		}

		// Send session state to client
		message, err := session.GetSessionState()
		if err != nil {
			fmt.Println("Error getting session state: ", err)
			return
		}
		err = session.SendMessageToSession(sessionId, message)
		if err != nil {
			fmt.Println("Error sending message to session: ", err)
			conn.Close()
			return
		}
		// Send Game Over Message
		if player.Type == models.ActivePlayerType {

			player1, err := session.GetPlayer(session.Game.Player1)
			if err != nil {
				fmt.Println(err)
				message := map[string]interface{}{
					"type":    models.GameOverMessageType,
					"message": "Game Over",
					"winner":  session.Game.GetWinner(),
				}
				session.State = models.SessionStateClosed
				session.BroadcastMessage(message)
				return
			}
			player2, err := session.GetPlayer(session.Game.Player2)
			if err != nil {
				fmt.Println(err)
				message := map[string]interface{}{
					"type":    models.GameOverMessageType,
					"message": "Game Over",
					"winner":  session.Game.GetWinner(),
				}
				session.State = models.SessionStateClosed
				session.BroadcastMessage(message)
				return
			}
			if utils.AbsInt64(player1.Time-player2.Time) > 15000 {
				session.State = models.SessionStateClosed
				message := map[string]interface{}{
					"type":    models.GameOverMessageType,
					"message": "Game Over",
					"winner":  session.Game.GetWinner(),
				}
				session.BroadcastMessage(message)
				return
			}
		}
	} else if session.State == models.SessionStateRematch && session.Rematch != nil {
		// Current Match is in rematch State
		if session.Rematch.Time-time.Now().UnixMilli() > 15000 {
			session.State = models.SessionStateClosed
			session.BroadcastSessionState()
		}
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
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": "You are not the active player"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		return
	}

	err = session.Game.MakeMove(player.ID, msg.Move)

	if err != nil {
		failedMove := map[string]interface{}{
			"type":    models.ErrorMessageType,
			"message": "Failed to make move",
		}
		conn.WriteJSON(failedMove)
		return
	}
	session.BroadcastGameState()
	if session.Game.IsGameOver() {
		session.ProcessGameOver()
		session.BroadcastGameOver()
	}
}

func handleChatMessage(conn *websocket.Conn, matchId string, sessionId string, msg models.Message) {
	fmt.Println("handleChatMessage", msg)
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
	fmt.Println("message made is past the get connection", message)
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
		errorMessage := map[string]string{"type": models.ErrorMessageType, "message": "You are not the active player"}
		fmt.Println(errorMessage)
		conn.WriteJSON(errorMessage)
		return
	}
	session.ProcessRematchRequest(player.ID)

	if session.Rematch != nil && session.Rematch.Player1Accepted && session.Rematch.Player2Accepted {
		// start the new game
		// TODO: Need to fix the function so we can have a random game player set
		session.StartNewGame(false)
		session.Rematch = nil
		session.BroadcastSessionState()
		session.BroadcastGameState()
	} else {
		message := map[string]interface{}{
			"type":    models.RematchMessageType,
			"message": "Rematch Request Received",
		}
		conn.WriteJSON(message)
		if session.Game.Player1 == player.ID {
			player2, err := session.GetPlayer(session.Game.Player2)
			if err != nil {
				fmt.Println(err)
				return
			}
			message := map[string]interface{}{
				"type":    models.RematchMessageType,
				"message": "Rematch Requested",
			}
			player2.Conn.WriteJSON(message)
		} else {
			player1, err := session.GetPlayer(session.Game.Player1)
			if err != nil {
				fmt.Println(err)
				return
			}
			message := map[string]interface{}{
				"type":    models.RematchMessageType,
				"message": "Rematch Requested",
			}
			player1.Conn.WriteJSON(message)
		}
	}

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
