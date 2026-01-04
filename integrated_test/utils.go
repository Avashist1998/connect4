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
func sendMessage(t *testing.T, conn *websocket.Conn, message map[string]interface{}, errorMessage string) {
	err := conn.WriteJSON(message)
	if err != nil {
		t.Fatalf(errorMessage+": %w", err)
	}
}

// readMessage reads a JSON message from the WebSocket connection with timeout
func readMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration, errorMessage string) map[string]interface{} {
	conn.SetReadDeadline(time.Now().Add(timeout))
	var msg map[string]interface{}
	err := conn.ReadJSON(&msg)
	if err != nil {
		t.Fatalf(errorMessage+": %w", err)
	}
	return msg
}

// waitForGameState waits for a "Game State" message from the WebSocket
func waitForGameState(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining < 100*time.Millisecond {
			remaining = 100 * time.Millisecond
		}
		msg := readMessage(t, conn, remaining, "Failed to read game state")
		if msg["type"] == "game_state" {
			return msg
		}
	}
	t.Fatalf("timeout waiting for game state")
	return nil
}
