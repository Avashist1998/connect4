package handlers

import (
	"4connect/internal/models"
	"4connect/internal/store"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func wsLiveV2WebSocketHandler(conn *websocket.Conn, r *http.Request) {
	playerId := ""
	liveManager := store.MakeLiveSessionManger()
	matchManager := store.MatchManagerFactory()

	// defer func() {
	// 	conn.Close()
	// 	go liveManager.RemovePlayer(matchId, playerId)
	// }()

	// Read JSON message from WebSocket
	var msg models.Message
	err := conn.ReadJSON(&msg)

	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}
	playerId = liveManager.AddPlayer(msg.MatchID, msg.Player, conn, matchManager)
	fmt.Printf("%s\n playerId :%s\n", msg, playerId)

}

func LiveWebSocketHandlerV2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	wsLiveV2WebSocketHandler(conn, r)
}
