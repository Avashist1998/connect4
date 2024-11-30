package handlers

import (
	"4connect/internal/models"
	"4connect/internal/store"
	"4connect/internal/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func wsLobbyMessageHandler(conn *websocket.Conn, r *http.Request) {

	playerID := ""
	lobby := store.LobbyFactory()
	manager := store.MatchManagerFactory()

	defer func() {
		conn.Close()
		if playerID != "" {
			lobby.RemovePlayer(playerID)
			log.Printf("Player %s disconnected and removed from lobby", playerID)
		}
	}()
	playerID = utils.GenerateId(10)
	fmt.Printf("New palyer ID was created %s\n", playerID)
	lobby.AddPlayer(playerID, conn, manager)
	fmt.Printf("New palyer ID was created %s\n", playerID)
	for {
		// Read JSON message from WebSocket
		var msg models.LobbyMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}
	}
}

func LobbyWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	fmt.Printf("we have a new connection\n")
	wsLobbyMessageHandler(conn, r)
}
