package integratedtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "http://0.0.0.0:9080"

func createMatch() (string, error) {
	// Create the request body
	requestData := map[string]interface{}{
		"GameType":    "live",
		"Player1":     "player1",
		"Player2":     "player2",
		"StartPlayer": "player1",
		"Level":       "",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", baseURL+"/", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	matchId, ok := response["match_id"].(string)
	if !ok {
		return "", fmt.Errorf("match_id not found in response")
	}

	return matchId, nil
}

func getSessionId() (string, error) {
	// Create the HTTP request
	req, err := http.NewRequest("POST", baseURL+"/session", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	sessionId, ok := response["session_id"].(string)
	if !ok {
		return "", fmt.Errorf("session_id not found in response")
	}

	return sessionId, nil
}

// connectWebSocket establishes a WebSocket connection to the live match endpoint
func connectWebSocket(matchId, sessionId string) (*websocket.Conn, error) {
	u := url.URL{
		Scheme:   "ws",
		Host:     "0.0.0.0:9080",
		Path:     "/ws/live",
		RawQuery: fmt.Sprintf("matchId=%s&sessionId=%s", matchId, sessionId),
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return conn, nil
}

// sendMessage sends a JSON message through the WebSocket connection
func sendMessage(conn *websocket.Conn, message map[string]interface{}) error {
	return conn.WriteJSON(message)
}

// readMessage reads a JSON message from the WebSocket connection with timeout
func readMessage(conn *websocket.Conn, timeout time.Duration) (map[string]interface{}, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	var msg map[string]interface{}
	err := conn.ReadJSON(&msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// waitForGameState waits for a "Game State" message from the WebSocket
func waitForGameState(conn *websocket.Conn, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining < 100*time.Millisecond {
			remaining = 100 * time.Millisecond
		}
		msg, err := readMessage(conn, remaining)
		if err != nil {
			return nil, err
		}
		if msg["message"] == "Game State" {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("timeout waiting for game state")
}

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
	sendMessage(conn, map[string]interface{}{"type": "join"})

	joinResp, err := readMessage(conn, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response: %v", err)
	}
	if joinResp["message"] != "joined session" {
		t.Fatalf("Expected 'joined session', got: %v", joinResp["message"])
	}

	sendMessage(conn, map[string]interface{}{"type": "ready", "name": "player1"})
	readyResp, err := readMessage(conn, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read ready response: %v", err)
	}

	if readyResp["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyResp["message"])
	}

	if readyResp["player"] != "player1" {
		t.Fatalf("Expected 'player1', got: %v", readyResp["player"])
	}

	if readyResp["type"] != "active" {
		t.Fatalf("Expected 'active', got: %v", readyResp["type"])
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

	sendMessage(conn, map[string]interface{}{"type": "join"})
	joinRespReconnect, err := readMessage(conn, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response: %v", err)
	}
	t.Logf("Join response after reconnect: %v", joinRespReconnect)
	if joinRespReconnect["message"] != "reconnected" {
		t.Fatalf("Expected 'reconnected', got: %v", joinRespReconnect["message"])
	}

	sendMessage(conn, map[string]interface{}{"type": "ready", "name": "player1"})
	readyRespReconnect, err := readMessage(conn, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to read ready response: %v", err)
	}

	if readyRespReconnect["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyRespReconnect["message"])
	}

	if readyRespReconnect["player"] != "player1" {
		t.Fatalf("Expected 'player1', got: %v", readyRespReconnect["player"])
	}

	if readyRespReconnect["type"] != "active" {
		t.Fatalf("Expected 'active', got: %v", readyRespReconnect["type"])
	}

	gameState, err := waitForGameState(conn, 5*time.Second)
	if err == nil {
		t.Fatalf("The game has started with only one player: %v", gameState)
	}
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
	if err := sendMessage(conn1, map[string]interface{}{"type": "join"}); err != nil {
		t.Fatalf("Failed to send join message for player 1: %v", err)
	}
	t.Logf("Player 1 sent join message")

	// Read join response - Player 1
	joinResp1, err := readMessage(conn1, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response for player 1: %v", err)
	}
	t.Logf("Player 1 join response: %v", joinResp1)

	// Send join message - Player 2
	if err := sendMessage(conn2, map[string]interface{}{"type": "join"}); err != nil {
		t.Fatalf("Failed to send join message for player 2: %v", err)
	}
	t.Logf("Player 2 sent join message")

	// Read join response - Player 2
	joinResp2, err := readMessage(conn2, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response for player 2: %v", err)
	}
	t.Logf("Player 2 join response: %v", joinResp2)

	// Send ready message - Player 1
	if err := sendMessage(conn1, map[string]interface{}{"type": "ready", "name": "player1"}); err != nil {
		t.Fatalf("Failed to send ready message for player 1: %v", err)
	}
	t.Logf("Player 1 sent ready message")

	// Read ready response - Player 1
	readyResp1, err := readMessage(conn1, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read ready response for player 1: %v", err)
	}
	t.Logf("Player 1 ready response: %v", readyResp1)

	// Send ready message - Player 2
	if err := sendMessage(conn2, map[string]interface{}{"type": "ready", "name": "player2"}); err != nil {
		t.Fatalf("Failed to send ready message for player 2: %v", err)
	}
	t.Logf("Player 2 sent ready message")

	// Read ready response - Player 2
	readyResp2, err := readMessage(conn2, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read ready response for player 2: %v", err)
	}
	t.Logf("Player 2 ready response: %v", readyResp2)

	// Wait for game state from both players
	t.Logf("Waiting for game state from both players...")
	gameState1, err := waitForGameState(conn1, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to get game state for player 1: %v", err)
	}
	t.Logf("Player 1 received game state: %v", gameState1)

	gameState2, err := waitForGameState(conn2, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to get game state for player 2: %v", err)
	}
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

	err = sendMessage(conn1, map[string]interface{}{"type": "move", "move": 0})
	if err != nil {
		t.Fatalf("Failed to send move message for player 1: %v", err)
	}
	t.Logf("Player 1 sent move message")
	gameState1AfterMove, err := waitForGameState(conn1, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to get game state for player 1: %v", err)
	}
	t.Logf("Player 1 received game state after move: %v", gameState1AfterMove)

	if reflect.DeepEqual(gameState1["board"], gameState1AfterMove["board"]) {
		t.Fatalf("Game board didn't change after move! Before: %v, After: %v", gameState1["board"], gameState1AfterMove["board"])
	}

	conn2, err = connectWebSocket(matchId, sessionId2)

	if err != nil {
		t.Fatalf("Failed to reconnect player 2: %v", err)
	}

	err = sendMessage(conn2, map[string]interface{}{"type": "join"})
	if err != nil {
		t.Fatalf("Failed to send join message for player 2: %v", err)
	}
	t.Logf("Player 2 sent join message")
	joinResp2Reconnect, err := readMessage(conn2, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response for player 2: %v", err)
	}
	t.Logf("Player 2 join response after reconnect: %v", joinResp2Reconnect)
	if joinResp2Reconnect["message"] != "reconnected" {
		t.Fatalf("Expected 'reconnected', got: %v", joinResp2Reconnect["message"])
	}

	err = sendMessage(conn2, map[string]interface{}{"type": "ready", "name": "player2"})
	if err != nil {
		t.Fatalf("Failed to send ready message for player 2: %v", err)
	}
	t.Logf("Player 2 sent ready message")
	readyResp2Reconnect, err := readMessage(conn2, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read ready response for player 2: %v", err)
	}
	t.Logf("Player 2 ready response after reconnect: %v", readyResp2Reconnect)
	if readyResp2Reconnect["message"] != "joined game" {
		t.Fatalf("Expected 'joined game', got: %v", readyResp2Reconnect["message"])
	}

	gameState2Reconnect, err := waitForGameState(conn2, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to get game state for player 2: %v", err)
	}
	t.Logf("Player 2 received game state after reconnect: %v", gameState2Reconnect)
	if reflect.DeepEqual(gameState2["board"], gameState2Reconnect["board"]) {
		t.Fatalf("Game board did not change after player 1 move and reconnect! Before: %v, After: %v", gameState2["board"], gameState2Reconnect["board"])
	}
	defer conn2.Close()
	t.Logf("Player 2 reconnected")

}
