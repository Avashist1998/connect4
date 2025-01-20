package game

import (
	"errors"
	"log"
	"math/rand"
)

type Game struct {
	moveCount   int
	Player1     string
	Player2     string
	winner      string
	moves       []int
	board       []int
}

func (game *Game) GetBoard() []int {

	boardCopy := make([]int, len(game.board))
	copy(boardCopy, game.board)
	return boardCopy
}

func NewGame(Player1 string, Player2 string) *Game {
	return &Game{
		Player1:     Player1,
		Player2:     Player2,
		moveCount:   0,
		moves:       []int{},
		board:       make([]int, 42),
		winner:      "",
	}
}

func (game *Game) isPlayerTurn(slot string) bool {
	if game.moveCount%2 == 0 {
		return slot == "RED"
	}
	return slot == "YELLOW"
}

func IsMoveValid(game *Game, move int) bool {

	if move < 0 {
		return false
	}

	if move > 6 {
		return false
	}

	for i := 0; i < 6; i++ {
		if game.board[i*7+move] == 0 {
			return true
		}
	}
	return false
}

func GetAxisCount(game *Game, index int, shift int) int {
	i := index
	col := index % 7
	count := 0
	for i < 42 && game.board[i] == game.board[index] {
		count += 1
		i += shift
		if (col == 0 && i%7 == 6) || (col == 6 && i%7 == 0) {
			break
		}
		col = i % 7
	}
	i = index
	for i > -1 && game.board[i] == game.board[index] {
		count += 1
		i += -shift
		if (col == 0 && i%7 == 6) || (col == 6 && i%7 == 0) {
			break
		}
		col = i % 7
	}
	return count - 1
}

func IsWinner(game *Game, index int) bool {
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
	count = GetAxisCount(game, index, 6)
	if count >= 4 {
		return true
	}
	// negative diagonal
	count = GetAxisCount(game, index, 8)
	return count >= 4
}

func (game *Game) GetWinner() string {
	return game.winner
}

func IsGameOver(game *Game) bool {
	if len(game.moves) == 42 {
		return true
	}
	return game.GetWinner() != ""
}

func MakeMove(game *Game, slot string, move int) error {
	if !game.isPlayerTurn(slot) {
		return errors.New("not your turn")
	}

	if !IsMoveValid(game, move) {
		return errors.New("cannot make this move")
	}

	if len(game.moves) == 42 || game.winner != "" {
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
		game.winner = slot
	}
	return nil
}

func ShowGame(game *Game) {
	for i := 5; i > -1; i-- {
		log.Printf("%v", game.board[i*7:i*7+7])
	}
}

func (game *Game) GetCurrPlayer() string {
	if game.isPlayerTurn("RED") {
		return game.Player1
	}
	return game.Player2
}

func (game *Game) GetMoves() []int {
	return game.moves
}

func UpdateNames(game *Game, player1 string, player2 string) {
	game.Player1 = player1
	game.Player2 = player2
}

func GetBotMove(game *Game, level string) int {
	randomIndex := int(rand.Float32() * 6)
	for !IsMoveValid(game, randomIndex) {
		randomIndex = int(rand.Float32() * 6)
	}
	return randomIndex
}
