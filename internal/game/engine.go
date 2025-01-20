package game

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func RunGame() {
	positionName := "RED"
	newGame := NewGame("Player1", "Player1")
	reader := bufio.NewReader(os.Stdin)
	player := "";
	for {
		player = newGame.GetCurrPlayer()
		fmt.Printf("%s move? : ", player)

		// Read and trim user input
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Convert input to an integer
		move, err := strconv.Atoi(input)
		if err != nil {
			println("Invalid input. Please enter a valid column number.")
			continue // Prompt the user again
		}

		// Attempt to make the move
		err = MakeMove(newGame, positionName, move)
		if err != nil {
			println(err.Error())
			continue // Prompt the user again
		}

		// Display the game board
		println("===============")
		ShowGame(newGame)
		println("===============")

		// Check for a winner
		if winner := newGame.GetWinner(); winner != "" {
			fmt.Printf("%s has won the game!\n", winner)
			break // Exit the loop if there's a winner
		}

		// Switch to the next player
		if positionName == "RED" {positionName = "YELLOW" } else { positionName = "RED" }
	}
}
