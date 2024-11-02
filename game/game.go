package game

import (
	"errors"
	"log"
)

type Game struct {
	moveCount   int
	player1     string
	player2     string
	startPlayer string
	winner      string
	moves       []int
	board       []int
}

func NewGame(player1 string, player2 string) *Game {
	return &Game{
		player1:     player1,
		player2:     player2,
		startPlayer: player1,
		moveCount:   0,
		moves:       []int{},
		board:       make([]int, 42),
		winner:      "",
	}
}

func IsPlayerTurn(game *Game, player string) bool {
	if game.moveCount%2 == 0 && player == game.startPlayer {
		return true
	}

	if game.moveCount%2 == 1 && player != game.startPlayer {
		return true
	}
	return false
}

func IsMoveValid(game *Game, move int) bool {
	for i := 0; i < 6; i++ {
		if game.board[i*7+move] == 0 {
			return true
		}
	}
	return false
}

func GetAxisCount(game *Game, index int, shift int) int {
	i := index
	count := 0
	for i < 42 && game.board[i] == game.board[index] {
		count += 1
		i += shift
	}
	i = index
	for i > -1 && game.board[i] == game.board[index] {
		count += 1
		i += -shift
	}
	return count - 1
}

func IsWinner(game *Game, index int) bool {
	println("The index is ", index)
	// vertical
	count := 0
	count = GetAxisCount(game, index, 7)
	if count >= 4 {
		return true
	}
	// horizontal
	count = GetAxisCount(game, index, 1)
	if count >= 4 {
		return true
	}
	// positive diagonal
	count = GetAxisCount(game, index, 4)
	if count >= 4 {
		return true
	}
	// negative diagonal
	count = GetAxisCount(game, index, 8)
	return count >= 4
}

func GetWinner(game *Game) string {
	return game.winner
}

func MakeMove(game *Game, player string, move int) error {
	if !IsPlayerTurn(game, player) {
		return errors.New("not your turn")
	}

	if !IsMoveValid(game, move) {
		return errors.New("cannot make this move")
	}

	if len(game.moves) == 42 {
		return errors.New("no more moves available")
	}

	if game.winner != "" {
		return errors.New("game is over")
	}
	val := 0
	if game.moveCount%2 == 0 {
		val = 1
	} else {
		val = -1
	}
	index := 0
	for i := 0; i < 6; i++ {
		if game.board[i*7+move] == 0 {
			game.board[i*7+move] = val
			index = i*7 + move
			break
		}
	}
	game.moveCount += 1
	game.moves = append(game.moves, move)
	if IsWinner(game, index) {
		game.winner = player
	}
	return nil
}

func ShowGame(game *Game) {
	for i := 5; i > -1; i-- {
		log.Printf("%v", game.board[i*7:i*7+7])
	}
}

func GetCurrPlayer(game *Game) string {
	if IsPlayerTurn(game, game.player1) {
		return game.player1
	}
	return game.player2
}
