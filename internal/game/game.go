package game

import (
	"errors"
	"log"
	"math/rand"
)

type Game struct {
	moveCount       int
	Player1         string
	Player2         string
	winner          string
	moves           []int
	board           []int
	currentPlayerID string
}

func (game *Game) GetBoard() []int {

	boardCopy := make([]int, len(game.board))
	copy(boardCopy, game.board)
	return boardCopy
}

func NewGame(playerID1 string, playerID2 string) (*Game, error) {

	if playerID1 == playerID2 {
		err := errors.New("players cannot be the same ID")
		return nil, err
	}
	return &Game{
		Player1:         playerID1,
		Player2:         playerID2,
		currentPlayerID: playerID1,
		moveCount:       0,
		moves:           []int{},
		board:           make([]int, 42),
		winner:          "",
	}, nil
}

func (game *Game) isPlayerTurn(playerID string) bool {
	if playerID == game.currentPlayerID {
		return true
	}
	return false
}

func (game *Game) isMoveValid(move int) bool {

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

func (game *Game) getAxisCount(index int, shift int) int {
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

func (game *Game) isWinner(index int) bool {
	// vertical
	count := 0
	count = game.getAxisCount(index, 7)
	if count >= 4 {
		return true
	}
	// horizontal
	count = game.getAxisCount(index, 1)
	if count >= 4 {
		return true
	}
	// positive diagonal
	count = game.getAxisCount(index, 6)
	if count >= 4 {
		return true
	}
	// negative diagonal
	count = game.getAxisCount(index, 8)
	return count >= 4
}

func (game *Game) GetWinner() string {
	return game.winner
}

func (game *Game) IsGameOver() bool {
	if len(game.moves) == 42 {
		return true
	}
	return game.GetWinner() != ""
}

func (game *Game) updateGameWinner(index int) {
	if !game.isWinner(index) {
		return
	}

	if game.currentPlayerID == game.Player2 {
		game.winner = game.Player2
		return
	}
	game.winner = game.Player1
}

func (game *Game) makeMove(move int) int {
	val := 0
	if game.currentPlayerID == game.Player1 {
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
	return index
}

func (game *Game) resetIndex(index int) {
	game.board[index] = 0
}

func (game *Game) getConnectionCount(index int) int {
	count := 0
	// vertical
	count += game.getAxisCount(index, 7)
	// horizontal
	count += game.getAxisCount(index, 1)
	// positive diagonal
	count += game.getAxisCount(index, 6)
	// negative diagonal
	count += game.getAxisCount(index, 8)
	return count
}

func (game *Game) MakeMove(playerID string, move int) error {
	if !game.isPlayerTurn(playerID) {
		return errors.New("not your turn")
	}

	if !game.isMoveValid(move) {
		return errors.New("cannot make this move")
	}

	if len(game.moves) == 42 || game.winner != "" {
		return errors.New("game is over")
	}
	index := game.makeMove(move)
	game.moveCount += 1
	game.moves = append(game.moves, move)
	game.updateGameWinner(index)
	if game.currentPlayerID == game.Player2 {
		game.currentPlayerID = game.Player1
	} else {
		game.currentPlayerID = game.Player2
	}
	return nil
}

func ShowGame(game *Game) {
	for i := 5; i > -1; i-- {
		log.Printf("%v", game.board[i*7:i*7+7])
	}
}

func (game *Game) GetCurrPlayer() string {
	return game.currentPlayerID
}

func (game *Game) GetCurrSlot() string {
	if game.moveCount%2 == 0 {
		return "RED"
	}
	return "YELLOW"
}

func (game *Game) GetMoves() []int {
	return game.moves
}

func (game *Game) getValidMoves() []int {
	var moves = []int{}

	for i := 0; i < 7; i++ {
		if game.isMoveValid(i) {
			moves = append(moves, i)
		}
	}
	return moves
}

func (game *Game) getRandomMove() int {
	moves := game.getValidMoves()
	return int(rand.Float32() * float32(len(moves)))
}

func (game *Game) simulateWinnerMove(move int) (bool, error) {

	if !game.isMoveValid(move) {
		return false, errors.New("invalid move")
	}

	index := game.makeMove(move)
	res := game.isWinner(index)
	game.resetIndex(index)
	return res, nil
}

func (game *Game) getMoveScore(move int) (int, error) {
	if !game.isMoveValid(move) {
		return -100, errors.New("invalid move")
	}
	score := 0
	index := game.makeMove(move)
	score += game.getConnectionCount(index)
	game.resetIndex(index)

	index = game.makeMove(move)
	score += game.getConnectionCount(index)
	game.resetIndex(index)

	return score, nil
}

func (game *Game) getSmartMove() int {

	moves := game.getValidMoves()
	for _, move := range moves {
		res, err := game.simulateWinnerMove(move)
		if err == nil && res {
			return move
		}
	}

	for _, move := range moves {
		res, err := game.simulateWinnerMove(move)
		if err == nil && res {
			return move
		}
	}

	bestMove := moves[0]
	bestScore := -4
	for _, move := range moves {
		score, err := game.getMoveScore(move)
		if err == nil {
			if score > bestScore {
				bestMove = move
				bestScore = score
			}
		}
	}
	return bestMove
}

func (game *Game) GetBotMove(level string) int {
	if level == "easy" {
		return game.getRandomMove()
	} else if level == "mid" {
		if rand.Float32() > 0.7 {
			return game.getSmartMove()
		}
		return game.getRandomMove()
	} else if level == "hard" {
		return game.getSmartMove()
	}
	return game.getRandomMove()
}
