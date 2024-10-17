package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pborman/getopt/v2"
)

var nextPlayer = map[string]string{
	"X": "O",
	"O": "X",
}

type Board struct {
	list []string
}

var wins = [][]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

func newBoard() *Board {
	b := &Board{}
	b.list = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	return b
}

func (b *Board) done() (bool, string) {
	for _, w := range wins {
		if b.list[w[0]] == b.list[w[1]] && b.list[w[1]] == b.list[w[2]] {
			if b.list[w[0]] == "X" {
				return true, "X"
			} else if b.list[w[0]] == "O" {
				return true, "O"
			}
		}
	}
	allXO := true
	for _, v := range b.list {
		if v != "X" && v != "O" {
			allXO = false
			break
		}
	}
	if allXO {
		return true, "tie"
	}
	return false, "going"
}

func (b *Board) String() string {
	var sb strings.Builder
	rows := [][]string{b.list[0:3], b.list[3:6], b.list[6:9]}
	for _, row := range rows {
		for _, cell := range row {
			if cell == "X" {
				sb.WriteString("\033[31m" + cell + "\033[0m | ") // Red color for X
			} else if cell == "O" {
				sb.WriteString("\033[34m" + cell + "\033[0m | ") // Blue color for O
			} else {
				sb.WriteString(cell + " | ")
			}
		}
		sb.WriteString("\n--------------\n")
	}
	return sb.String()
}

func (b *Board) emptySpots() []int {
	var emptyIndices []int
	for _, v := range b.list {
		if _, err := strconv.Atoi(v); err == nil {
			i, _ := strconv.Atoi(v)
			emptyIndices = append(emptyIndices, i)
		}
	}
	return emptyIndices
}

type Move struct {
	score int
	idx   int
}

type Moves []Move

func (m Moves) Len() int           { return len(m) }
func (m Moves) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Moves) Less(i, j int) bool { return m[i].score < m[j].score }

type Game struct {
	currentPlayer string
	board         *Board
	aiPlayer      string
	difficulty    int
}

func newGame(aiPlayer string, difficulty int) *Game {
	game := &Game{
		board:         newBoard(),
		currentPlayer: "X",
		aiPlayer:      aiPlayer,
		difficulty:    difficulty,
	}
	return game
}

func (g *Game) changePlayer() {
	g.currentPlayer = nextPlayer[g.currentPlayer]
}

func (g *Game) getBestMove(board *Board, player string) Move {
	done, winner := board.done()
	if done {
		if winner == g.aiPlayer {
			return Move{score: 10, idx: 0}
		} else if winner != "tie" {
			return Move{score: -10, idx: 0}
		} else {
			return Move{score: 0, idx: 0}
		}
	}

	emptySpots := board.emptySpots()
	var moves Moves
	for _, idx := range emptySpots {
		newBoard := newBoard()
		newBoard.list = make([]string, len(board.list))
		copy(newBoard.list, board.list)
		newBoard.list[idx] = player
		score := g.getBestMove(newBoard, nextPlayer[player]).score
		moves = append(moves, Move{score: score, idx: idx})
	}

	if player == g.aiPlayer {
		sort.Sort(sort.Reverse(moves))
		return moves[0]
	} else {
		sort.Sort(moves)
		return moves[0]
	}
}

func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("clear")
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		fmt.Println("Unsupported operating system for clearing the screen.")
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (g *Game) startGame() {
	for {
		clearScreen() // Clear the screen before printing a new move

		fmt.Println(g.board)

		g.changePlayer() // Call changePlayer *before* checking whose turn it is

		if g.aiPlayer == g.currentPlayer {
			emptySpots := g.board.emptySpots()
			if len(emptySpots) > g.difficulty { // Corrected difficulty logic
				fmt.Println("RANDOM GUESS")
				g.board.list[emptySpots[rand.Intn(len(emptySpots))]] = g.aiPlayer
			} else {
				fmt.Println("AI MOVE..")
				move := g.getBestMove(g.board, g.aiPlayer)
				g.board.list[move.idx] = g.aiPlayer
			}
		} else {
			fmt.Print("Enter move (0-8): ")
			var moveStr string
			fmt.Scanln(&moveStr)
			move, err := strconv.Atoi(moveStr)
			if err != nil || move < 0 || move > 8 || g.board.list[move] == "X" || g.board.list[move] == "O" {
				fmt.Println("Invalid move. Please try again.")
				continue // Go back to the beginning of the loop
			}
			g.board.list[move] = g.currentPlayer
		}

		done, winner := g.board.done()
		if done {
			clearScreen() // Clear the screen before printing the final result

			fmt.Println(g.board)
			if winner == "tie" {
				fmt.Println("TIE")
			} else {
				fmt.Println("WINNER IS :", winner)
			}
			break
		}
	}
}

func writeHelp() {
	fmt.Println(`
TicTacToe 0.1.0 (MinMax version)
Allowed arguments:
  -h | --help         : show help
  -a | --ai           : AI player [X or O]
  -l | --difficulty   : difficulty level (0-8)`)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	helpFlag := getopt.BoolLong("help", 'h', "show help")
	aiPlayerFlag := getopt.StringLong("ai", 'a', "", "AI player [X or O]")
	difficultyFlag := getopt.IntLong("difficulty", 'l', 9, "difficulty level (0-8)")

	getopt.Parse()

	if *helpFlag {
		writeHelp()
		os.Exit(0)
	}

	g := newGame(*aiPlayerFlag, *difficultyFlag)
	g.startGame()
}
