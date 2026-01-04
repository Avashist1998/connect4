package integratedtest

import (
	"reflect"
	"testing"
	"time"
)

func TestRematchBothPlayersAccept(t *testing.T) {
	matchId, err := createMatch()
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}
	t.Logf("Created match with ID: %s", matchId)

	sessionId1, err := getSessionId()
	if err != nil {
		t.Fatalf("Failed to get session ID for player 1: %v", err)
	}
	t.Logf("Got session ID for player 1: %s", sessionId1)

	sessionId2, err := getSessionId()
	if err != nil {
		t.Fatalf("Failed to get session ID for player 2: %v", err)
	}
	t.Logf("Got session ID for player 2: %s", sessionId2)

	conn1, err := connectWebSocket(matchId, sessionId1)
	if err != nil {
		t.Fatalf("Failed to connect player 1: %v", err)
	}
	defer conn1.Close()
	conn2, err := connectWebSocket(matchId, sessionId2)
	if err != nil {
		t.Fatalf("Failed to connect player 2: %v", err)
	}
	defer conn2.Close()

	var message1 map[string]interface{}
	var message2 map[string]interface{}
	sendMessage(t, conn1, map[string]interface{}{"type": "join"}, "Failed to send join message for player 1")
	sendMessage(t, conn2, map[string]interface{}{"type": "join"}, "Failed to send join message for player 2")
	joinResp1 := readMessage(t, conn1, 5*time.Second, "Failed to read join response for player 1")
	t.Logf("Player 1 join type: %s, response: %s", joinResp1["type"], joinResp1["message"])
	joinResp2 := readMessage(t, conn2, 5*time.Second, "Failed to read join response for player 2")
	t.Logf("Player 2 join type: %s, response: %s", joinResp2["type"], joinResp2["message"])

	sendMessage(t, conn1, map[string]interface{}{"type": "ready", "name": "player1"}, "Failed to send ready message for player 1")
	time.Sleep(50 * time.Millisecond)
	sendMessage(t, conn2, map[string]interface{}{"type": "ready", "name": "player2"}, "Failed to send ready message for player 2")
	readyResp1 := readMessage(t, conn1, 5*time.Second, "Failed to read ready response for player 1")
	t.Logf("Player 1 ready type: %s, response: %v ", readyResp1["type"], readyResp1)
	readyResp2 := readMessage(t, conn2, 5*time.Second, "Failed to read ready response for player 2")
	t.Logf("Player 2 ready type: %s, response: %v", readyResp2["type"], readyResp2)
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if message1["type"] != "game_state" || message2["type"] != "game_state" {
		t.Fatalf("Expected game_state, got: %v for player 1 and %v for player 2", message1["type"], message2["type"])
	}
	if message1["you"] != "player1" || message2["you"] != "player2" {
		t.Fatalf("Expected you to be player1 for player 1 and player2 for player 2, got: %v for player 1 and %v for player 2", message1["you"], message2["you"])
	}
	t.Logf("Game is ready to recieveing moves")
	// Game Moves
	sendMessage(t, conn1, map[string]interface{}{"type": "move", "move": 0}, "Failed to send move message for player 1")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn2, map[string]interface{}{"type": "move", "move": 1}, "Failed to send move message for player 2")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn1, map[string]interface{}{"type": "move", "move": 0}, "Failed to send move message for player 1")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn2, map[string]interface{}{"type": "move", "move": 1}, "Failed to send move message for player 2")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn1, map[string]interface{}{"type": "move", "move": 0}, "Failed to send move message for player 1")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn2, map[string]interface{}{"type": "move", "move": 1}, "Failed to send move message for player 2")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	sendMessage(t, conn1, map[string]interface{}{"type": "move", "move": 0}, "Failed to send move message for player 1")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if !reflect.DeepEqual(message1["board"], message2["board"]) {
		t.Fatalf("Expected game boards to be equal, got: %v for player 1 and %v for player 2", message1["board"], message2["board"])
	}

	if message1["winner"] != "player1" || message2["winner"] != "player1" {
		t.Fatalf("Expected player1 to win, got: %v for player 1 and %v for player 2", message1["winner"], message2["winner"])
	}

	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if message1["type"] != "game_over" || message2["type"] != "game_over" {
		t.Fatalf("Expected game_over, got: %v for player 1 and %v for player 2", message1["type"], message2["type"])
	}
	// Rematch Validation
	sendMessage(t, conn1, map[string]interface{}{"type": "rematch"}, "Failed to send rematch message for player 1")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	if message1["type"] != "rematch" || message2["type"] != "rematch" {
		t.Fatalf("Expected rematch, got: %v for player 1 and %v for player 2", message1["type"], message2["type"])
	}
	sendMessage(t, conn2, map[string]interface{}{"type": "rematch"}, "Failed to send rematch message for player 2")
	message1 = readMessage(t, conn1, 5*time.Second, "Failed to read game state for player 1")
	message2 = readMessage(t, conn2, 5*time.Second, "Failed to read game state for player 2")
	t.Logf("Player 1 rematch type: %s, response: %v", message1["type"], message1)
	t.Logf("Player 2 rematch type: %s, response: %v", message2["type"], message2)
}
