package integratedtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const baseURL = "http://0.0.0.0:9080"

func createMatch() (string, error) {
	// Create the request body
	requestData := map[string]string{
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
func sendMessage(conn *websocket.Conn, message interface{}) error {
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
	//
	// We create a match and get the match id
	// We get a session id
	// Connect to the server to the match using the websocket using /ws/live?matchId=matchId&sessionId=sessionId
	// Send {"type": "join"} then {"type": "ready", "name": "name"}
	// Then break the conneciton
	// Then reconnect to see if the session is connected

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
	sendMessage(conn, map[string]string{"type": "join"})

	joinResp, err := readMessage(conn, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response: %v", err)
	}
	if joinResp["message"] != "joined session" {
		t.Fatalf("Expected 'joined session', got: %v", joinResp["message"])
	}

	sendMessage(conn, map[string]string{"type": "ready", "name": "player1"})
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

	sendMessage(conn, map[string]string{"type": "join"})
	joinRespReconnect, err := readMessage(conn, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to read join response: %v", err)
	}
	t.Logf("Join response after reconnect: %v", joinRespReconnect)
	if joinRespReconnect["message"] != "reconnected" {
		t.Fatalf("Expected 'reconnected', got: %v", joinRespReconnect["message"])
	}

	sendMessage(conn, map[string]string{"type": "ready", "name": "player1"})
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

// func TestReconnectTwoPlayers(t *testing.T) {
// 	//
// 	// We create a match and get the match id
// 	// We get a session id - Player 1
// 	// We get another session id - Player 2
// 	// Connect to the server to the match using the websocket using /ws/live?matchId=matchId&sessionId=sessionId - Player 1
// 	// Connect to the server to the match using the websocket using /ws/live?matchId=matchId&sessionId=sessionId - Player 2
// 	// Send {"type": "join"} then {"type": "ready", "name": "player1"}
// 	// Send {"type": "join"} then {"type": "ready", "name": "player2"}

// 	// After player 1 we should get the game state from both players make sure they are the same
// 	// Then break the connection - Player 1
// 	// Then reconnect to see if the session is connected - Player 1
// 	// Then check if the game state is the same as the first move

// 	// Create a match
// 	matchId, err := createMatch()
// 	if err != nil {
// 		t.Fatalf("Failed to create match: %v", err)
// 	}
// 	t.Logf("Created match with ID: %s", matchId)

// 	// Get session id - Player 1
// 	sessionId1, err := getSessionId()
// 	if err != nil {
// 		t.Fatalf("Failed to get session ID for player 1: %v", err)
// 	}
// 	t.Logf("Got session ID for player 1: %s", sessionId1)

// 	// Get session id - Player 2
// 	sessionId2, err := getSessionId()
// 	if err != nil {
// 		t.Fatalf("Failed to get session ID for player 2: %v", err)
// 	}
// 	t.Logf("Got session ID for player 2: %s", sessionId2)

// 	// Connect Player 1
// 	conn1, err := connectWebSocket(matchId, sessionId1)
// 	if err != nil {
// 		t.Fatalf("Failed to connect player 1: %v", err)
// 	}
// 	defer conn1.Close()
// 	t.Logf("Player 1 connected")

// 	// Connect Player 2
// 	conn2, err := connectWebSocket(matchId, sessionId2)
// 	if err != nil {
// 		t.Fatalf("Failed to connect player 2: %v", err)
// 	}
// 	defer conn2.Close()
// 	t.Logf("Player 2 connected")

// 	// Send join message - Player 1
// 	if err := sendMessage(conn1, "join", ""); err != nil {
// 		t.Fatalf("Failed to send join message for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 sent join message")

// 	// Read join response - Player 1
// 	joinResp1, err := readMessage(conn1, 5*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to read join response for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 join response: %v", joinResp1)

// 	// Send join message - Player 2
// 	if err := sendMessage(conn2, "join", ""); err != nil {
// 		t.Fatalf("Failed to send join message for player 2: %v", err)
// 	}
// 	t.Logf("Player 2 sent join message")

// 	// Read join response - Player 2
// 	joinResp2, err := readMessage(conn2, 5*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to read join response for player 2: %v", err)
// 	}
// 	t.Logf("Player 2 join response: %v", joinResp2)

// 	// Send ready message - Player 1
// 	if err := sendMessage(conn1, "ready", "player1"); err != nil {
// 		t.Fatalf("Failed to send ready message for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 sent ready message")

// 	// Read ready response - Player 1
// 	readyResp1, err := readMessage(conn1, 5*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to read ready response for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 ready response: %v", readyResp1)

// 	// Send ready message - Player 2
// 	if err := sendMessage(conn2, "ready", "player2"); err != nil {
// 		t.Fatalf("Failed to send ready message for player 2: %v", err)
// 	}
// 	t.Logf("Player 2 sent ready message")

// 	// Read ready response - Player 2
// 	readyResp2, err := readMessage(conn2, 5*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to read ready response for player 2: %v", err)
// 	}
// 	t.Logf("Player 2 ready response: %v", readyResp2)

// 	// Wait for game state from both players
// 	t.Logf("Waiting for game state from both players...")
// 	gameState1, err := waitForGameState(conn1, 10*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to get game state for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 received game state: %v", gameState1)

// 	gameState2, err := waitForGameState(conn2, 10*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to get game state for player 2: %v", err)
// 	}
// 	t.Logf("Player 2 received game state: %v", gameState2)

// 	// Verify both players have the same game state
// 	if gameState1["board"] != gameState2["board"] {
// 		t.Errorf("Game boards don't match! Player 1: %v, Player 2: %v", gameState1["board"], gameState2["board"])
// 	}
// 	if gameState1["currPlayer"] != gameState2["currPlayer"] {
// 		t.Errorf("Current players don't match! Player 1: %v, Player 2: %v", gameState1["currPlayer"], gameState2["currPlayer"])
// 	}
// 	if gameState1["state"] != gameState2["state"] {
// 		t.Errorf("Game states don't match! Player 1: %v, Player 2: %v", gameState1["state"], gameState2["state"])
// 	}
// 	t.Logf("Game states match between both players")

// 	// Store the initial game state for comparison after reconnect
// 	initialBoard := gameState1["board"]
// 	initialCurrPlayer := gameState1["currPlayer"]
// 	initialState := gameState1["state"]

// 	// Break the connection - Player 1
// 	t.Logf("Disconnecting player 1...")
// 	conn1.Close()

// 	// Wait a bit for the connection to fully close
// 	time.Sleep(1 * time.Second)

// 	// Reconnect Player 1
// 	t.Logf("Reconnecting player 1...")
// 	conn1Reconnect, err := connectWebSocket(matchId, sessionId1)
// 	if err != nil {
// 		t.Fatalf("Failed to reconnect player 1: %v", err)
// 	}
// 	defer conn1Reconnect.Close()
// 	t.Logf("Player 1 reconnected")

// 	// Send join message again - Player 1 (reconnect)
// 	if err := sendMessage(conn1Reconnect, "join", ""); err != nil {
// 		t.Fatalf("Failed to send join message for reconnected player 1: %v", err)
// 	}

// 	// Read reconnect response - Player 1
// 	reconnectResp, err := readMessage(conn1Reconnect, 5*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to read reconnect response for player 1: %v", err)
// 	}
// 	t.Logf("Player 1 reconnect response: %v", reconnectResp)

// 	// Verify reconnection message
// 	if reconnectResp["message"] != "Reconnected to old session" {
// 		t.Errorf("Expected 'Reconnected to old session', got: %v", reconnectResp["message"])
// 	}

// 	// Wait for game state after reconnect
// 	t.Logf("Waiting for game state after reconnect...")
// 	reconnectGameState, err := waitForGameState(conn1Reconnect, 10*time.Second)
// 	if err != nil {
// 		t.Fatalf("Failed to get game state after reconnect: %v", err)
// 	}
// 	t.Logf("Player 1 received game state after reconnect: %v", reconnectGameState)

// 	// Verify the game state is the same as before disconnection
// 	if reconnectGameState["board"] != initialBoard {
// 		t.Errorf("Game board changed after reconnect! Before: %v, After: %v", initialBoard, reconnectGameState["board"])
// 	}
// 	if reconnectGameState["currPlayer"] != initialCurrPlayer {
// 		t.Errorf("Current player changed after reconnect! Before: %v, After: %v", initialCurrPlayer, reconnectGameState["currPlayer"])
// 	}
// 	if reconnectGameState["state"] != initialState {
// 		t.Errorf("Game state changed after reconnect! Before: %v, After: %v", initialState, reconnectGameState["state"])
// 	}
// 	t.Logf("Game state matches after reconnect - test passed!")
// }
