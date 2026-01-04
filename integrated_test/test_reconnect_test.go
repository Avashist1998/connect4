package integratedtest

import (
	"reflect"
	"testing"
	"time"
)

func TestReconnect(t *testing.T) {
	// Create a match
	matchId, err := createMatch()
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}
	t.Logf("Created match with ID: %s", matchId)

	// Get a session id
	sessionId, err := getSessionId()
	if err != nil {
		t.Fatalf("Failed to get session ID: %v", err)
	}
	t.Logf("Got session ID: %s", sessionId)

	// Connect to the server to the match using the websocket using /ws/live?matchId=matchId&sessionId=sessionId
	conn, err := connectWebSocket(matchId, sessionId)
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	sendMessage(t, conn, map[string]interface{}{"type": "join"}, "Failed to send join message")

	joinResp := readMessage(t, conn, 5*time.Second, "Failed to read join response")
	if joinResp["message"] != "joined session" {
		t.Fatalf("Expected 'joined session', got: %v", joinResp["message"])
	}

	sendMessage(t, conn, map[string]interface{}{"type": "ready", "name": "player1"}, "Failed to send ready message for player 1")
	readyResp := readMessage(t, conn, 5*time.Second, "Failed to read ready response")

	if readyResp["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyResp["message"])
	}

	if readyResp["player"] != "player1" {
		t.Fatalf("Expected 'player1', got: %v", readyResp["player"])
	}

	if readyResp["player_type"] != "active" {
		t.Fatalf("Expected 'active', got: %v", readyResp["player_type"])
	}

	t.Logf("Player 1 ready response: %v", readyResp)
	t.Logf("Closing connection...")
	conn.Close()
	time.Sleep(1 * time.Second)
	t.Logf("Reconnecting...")
	conn, err = connectWebSocket(matchId, sessionId)
	if err != nil {
		t.Fatalf("Failed to reconnect to websocket: %v", err)
	}
	defer conn.Close()
	t.Logf("Reconnected to websocket")

	sendMessage(t, conn, map[string]interface{}{"type": "join"}, "Failed to send join message after reconnect")
	joinRespReconnect := readMessage(t, conn, 5*time.Second, "Failed to read join response after reconnect")
	t.Logf("Join response after reconnect: %v", joinRespReconnect)
	if joinRespReconnect["message"] != "reconnected" {
		t.Fatalf("Expected 'reconnected', got: %v", joinRespReconnect["message"])
	}

	sendMessage(t, conn, map[string]interface{}{"type": "ready", "name": "player1"}, "Failed to send ready message after reconnect")
	readyRespReconnect := readMessage(t, conn, 2*time.Second, "Failed to read ready response after reconnect")
	if readyRespReconnect["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyRespReconnect["message"])
	}

	if readyRespReconnect["player"] != "player1" {
		t.Fatalf("Expected 'player1', got: %v", readyRespReconnect["player"])
	}

	if readyRespReconnect["player_type"] != "active" {
		t.Fatalf("Expected 'active', got: %v", readyRespReconnect["player_type"])
	}

	sessionState := readMessage(t, conn, 5*time.Second, "Failed to get the session state after reconnect")
	t.Logf("Session state: %v", sessionState)
	if sessionState["player1Status"] != "live" {
		t.Fatalf("Expected 'live', got: %v", sessionState["state"])
	}
	if sessionState["player2Status"] != "disconnected" {
		t.Fatalf("Expected 'disconnected', got: %v", sessionState["player2Status"])
	}
	t.Logf("Passed the Game has not started yet check")

}

func TestReconnectTwoPlayers(t *testing.T) {
	// Create a match
	matchId, err := createMatch()
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}
	t.Logf("Created match with ID: %s", matchId)

	// Get session id - Player 1
	sessionId1, err := getSessionId()
	if err != nil {
		t.Fatalf("Failed to get session ID for player 1: %v", err)
	}
	t.Logf("Got session ID for player 1: %s", sessionId1)

	// Get session id - Player 2
	sessionId2, err := getSessionId()
	if err != nil {
		t.Fatalf("Failed to get session ID for player 2: %v", err)
	}
	t.Logf("Got session ID for player 2: %s", sessionId2)

	// Connect Player 1
	conn1, err := connectWebSocket(matchId, sessionId1)
	if err != nil {
		t.Fatalf("Failed to connect player 1: %v", err)
	}
	defer conn1.Close()
	t.Logf("Player 1 connected")

	// Connect Player 2
	conn2, err := connectWebSocket(matchId, sessionId2)
	if err != nil {
		t.Fatalf("Failed to connect player 2: %v", err)
	}
	defer conn2.Close()
	t.Logf("Player 2 connected")

	// Send join message - Player 1
	sendMessage(t, conn1, map[string]interface{}{"type": "join"}, "Failed to send join message for player 1")
	t.Logf("Player 1 sent join message")

	// Read join response - Player 1
	joinResp1 := readMessage(t, conn1, 2*time.Second, "Failed to read join response for player 1")
	t.Logf("Player 1 join response: %v", joinResp1)

	// Send join message - Player 2
	sendMessage(t, conn2, map[string]interface{}{"type": "join"}, "Failed to send join message for player 2")
	t.Logf("Player 2 sent join message")

	// Read join response - Player 2
	joinResp2 := readMessage(t, conn2, 2*time.Second, "Failed to read join response for player 2")
	t.Logf("Player 2 join response: %v", joinResp2)

	// Send ready message - Player 1
	sendMessage(t, conn1, map[string]interface{}{"type": "ready", "name": "player1"}, "Failed to send ready message for player 1")
	t.Logf("Player 1 sent ready message")

	// Read ready response - Player 1
	readyResp1 := readMessage(t, conn1, 5*time.Second, "Failed to read ready response for player 1")
	t.Logf("Player 1 ready response: %v", readyResp1)

	// Send ready message - Player 2
	sendMessage(t, conn2, map[string]interface{}{"type": "ready", "name": "player2"}, "Failed to send ready message for player 2")
	t.Logf("Player 2 sent ready message")

	// Read ready response - Player 2
	readyResp2 := readMessage(t, conn2, 5*time.Second, "Failed to read ready response for player 2")
	t.Logf("Player 2 ready response: %v", readyResp2)

	// Wait for game state from both players
	t.Logf("Waiting for game state from both players...")
	gameState1 := waitForGameState(t, conn1, 10*time.Second)
	t.Logf("Player 1 received game state: %v", gameState1)

	gameState2 := waitForGameState(t, conn2, 10*time.Second)
	t.Logf("Player 2 received game state: %v", gameState2)

	// Verify both players have the same game state
	if !reflect.DeepEqual(gameState1["board"], gameState2["board"]) {
		t.Errorf("Game boards don't match! Player 1: %v, Player 2: %v", gameState1["board"], gameState2["board"])
	}
	if gameState1["currPlayer"] != gameState2["currPlayer"] {
		t.Errorf("Current players don't match! Player 1: %v, Player 2: %v", gameState1["currPlayer"], gameState2["currPlayer"])
	}
	if gameState1["you"] != "player1" {
		t.Errorf("You don't match! Player 1: %v, Player 2: %v", gameState1["you"], gameState2["you"])
	}
	if gameState2["you"] != "player2" {
		t.Errorf("You don't match! Player 1: %v, Player 2: %v", gameState1["you"], gameState2["you"])
	}
	t.Logf("Game states match between both players %v", gameState1)

	// Break the connection - Player 2
	conn2.Close()
	time.Sleep(500 * time.Millisecond)

	sendMessage(t, conn1, map[string]interface{}{"type": "move", "move": 0}, "Failed to send move message for player 1")
	t.Logf("Player 1 sent move message")
	gameState1AfterMove := waitForGameState(t, conn1, 5*time.Second)
	t.Logf("Player 1 received game state after move: %v", gameState1AfterMove)

	if reflect.DeepEqual(gameState1["board"], gameState1AfterMove["board"]) {
		t.Fatalf("Game board didn't change after move! Before: %v, After: %v", gameState1["board"], gameState1AfterMove["board"])
	}

	conn2, err = connectWebSocket(matchId, sessionId2)

	if err != nil {
		t.Fatalf("Failed to reconnect player 2: %v", err)
	}

	sendMessage(t, conn2, map[string]interface{}{"type": "join"}, "Failed to send join message for player 2")
	t.Logf("Player 2 sent join message")
	joinResp2Reconnect := readMessage(t, conn2, 2*time.Second, "Failed to read join response for player 2")
	t.Logf("Player 2 join response after reconnect: %v", joinResp2Reconnect)
	if joinResp2Reconnect["message"] != "reconnected" {
		t.Fatalf("Expected 'reconnected', got: %v", joinResp2Reconnect["message"])
	}

	sendMessage(t, conn2, map[string]interface{}{"type": "ready", "name": "player2"}, "Failed to send ready message for player 2")
	t.Logf("Player 2 sent ready message")
	readyResp2Reconnect := readMessage(t, conn2, 5*time.Second, "Failed to read ready response for player 2")
	t.Logf("Player 2 ready response after reconnect: %v", readyResp2Reconnect)
	if readyResp2Reconnect["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyResp2Reconnect["message"])
	}

	gameState2Reconnect := waitForGameState(t, conn2, 5*time.Second)
	t.Logf("Player 2 received game state after reconnect: %v", gameState2Reconnect)
	if reflect.DeepEqual(gameState2["board"], gameState2Reconnect["board"]) {
		t.Fatalf("Game board did not change after player 1 move and reconnect! Before: %v, After: %v", gameState2["board"], gameState2Reconnect["board"])
	}
	defer conn2.Close()
	t.Logf("Player 2 reconnected")

}
