package game

import (
	"testing"
)

/*
TODO:
Most Important (Core Game Logic & Rules)

    MakeMove(game *Game, slot string, move int) error
        Ensures moves follow game rules, including turn order and valid moves.
        Essential for gameplay correctness.

    isWinner(index int) bool
        Determines if a player has won.
        A bug here could break the core game functionality.

    isMoveValid(move int) bool
        Prevents illegal moves from being made.

    updateGameWinner(index int)
        Ensures the correct player is marked as the winner.

    isPlayerTurn(slot string) bool
        Ensures correct player turn enforcement.

    GetWinner() string
        Returns the correct winner when the game ends.
*/

func TestGameInitialization(t *testing.T) {
	// Example test case

	g := NewGame("player1", "player2")
	
	if g.Player1 != "player1"  {
		t.Errorf("Expected %s, got %s", "player1" , g.Player1)
	}

	if g.Player2 != "player2" {
		t.Errorf("Expected %s, got %s", "player2" , g.Player2)
	}
}


