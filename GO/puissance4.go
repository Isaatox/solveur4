package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	ROWS = 6
	COLS = 7
)

type Player int
type Cell int
type Board [][]Cell

const (
	EMPTY Cell   = 0
	P1    Player = 1
	P2    Player = 2
)

type GameOutcome string

const (
	WIN_P1 GameOutcome = "WIN_P1"
	WIN_P2 GameOutcome = "WIN_P2"
	DRAW   GameOutcome = "DRAW"
)

type GameNode struct {
	Move     int
	Outcome  GameOutcome
	Children []GameNode
}

func createBoard() Board {
	board := make(Board, ROWS)
	for i := range board {
		board[i] = make([]Cell, COLS)
	}
	return board
}

func cloneBoard(board Board) Board {
	newBoard := make(Board, ROWS)
	for i := range board {
		newBoard[i] = make([]Cell, COLS)
		copy(newBoard[i], board[i])
	}
	return newBoard
}

func getAvailableRow(board Board, col int) int {
	for row := ROWS - 1; row >= 0; row-- {
		if board[row][col] == EMPTY {
			return row
		}
	}
	return -1
}

func dropPiece(board Board, row, col int, player Player) {
	board[row][col] = Cell(player)
}

func isWinningMove(board Board, player Player) bool {
	directions := [][]int{
		{0, 1}, {1, 0}, {1, 1}, {1, -1},
	}

	for r := 0; r < ROWS; r++ {
		for c := 0; c < COLS; c++ {
			if board[r][c] != Cell(player) {
				continue
			}
			for _, dir := range directions {
				count := 0
				for i := 0; i < 4; i++ {
					nr := r + dir[0]*i
					nc := c + dir[1]*i
					if nr >= 0 && nr < ROWS && nc >= 0 && nc < COLS && board[nr][nc] == Cell(player) {
						count++
					}
				}
				if count == 4 {
					return true
				}
			}
		}
	}
	return false
}

func getValidColumns(board Board) []int {
	cols := []int{}
	for c := 0; c < COLS; c++ {
		if getAvailableRow(board, c) != -1 {
			cols = append(cols, c)
		}
	}
	return cols
}

func exploreGameTree(board Board, currentPlayer Player, depth int) []GameNode {
	opponent := P2
	if currentPlayer == P2 {
		opponent = P1
	}
	validCols := prioritizeCenter(getValidColumns(board))
	results := []GameNode{}

	if depth == 0 || len(validCols) == 0 {
		return []GameNode{{Move: -1, Outcome: DRAW}}
	}

	for _, col := range validCols {
		row := getAvailableRow(board, col)
		if row == -1 {
			continue
		}
		newBoard := cloneBoard(board)
		dropPiece(newBoard, row, col, currentPlayer)

		if isWinningMove(newBoard, currentPlayer) {
			outcome := WIN_P1
			if currentPlayer == P2 {
				outcome = WIN_P2
			}
			results = append(results, GameNode{Move: col, Outcome: outcome})
		} else {
			children := exploreGameTree(newBoard, opponent, depth-1)
			results = append(results, GameNode{
				Move:     col,
				Outcome:  summarizeOutcomes(children, currentPlayer),
				Children: children,
			})
		}
	}

	return results
}

func summarizeOutcomes(nodes []GameNode, player Player) GameOutcome {
	win := WIN_P1
	lose := WIN_P2
	if player == P2 {
		win = WIN_P2
		lose = WIN_P1
	}

	allLose := true
	for _, node := range nodes {
		if node.Outcome == win {
			return win
		}
		if node.Outcome != lose {
			allLose = false
		}
	}
	if allLose {
		return lose
	}
	return DRAW
}

func prioritizeCenter(cols []int) []int {
	center := COLS / 2
	sorted := make([]int, len(cols))
	copy(sorted, cols)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if abs(center-sorted[j]) < abs(center-sorted[i]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func findShortestWinningPath(nodes []GameNode, target GameOutcome) []int {
	var shortest []int
	var dfs func([]int, []GameNode, int, Player)

	dfs = func(path []int, children []GameNode, depth int, currentPlayer Player) {
		for _, node := range children {
			newPath := append([]int{}, path...)
			newPath = append(newPath, node.Move)

			if node.Outcome == target && len(node.Children) == 0 {
				if shortest == nil || len(newPath) < len(shortest) {
					shortest = newPath
				}
			} else if len(node.Children) > 0 {
				nextPlayer := P2
				if currentPlayer == P2 {
					nextPlayer = P1
				}
				dfs(newPath, node.Children, depth+1, nextPlayer)
			}
		}
	}

	dfs([]int{}, nodes, 0, P1)
	return shortest
}

func simulateGame(board Board, moves int) Board {
	currentPlayer := P1
	rand.Seed(time.Now().UnixNano())

	for moves > 0 {
		validCols := getValidColumns(board)
		if len(validCols) == 0 {
			break
		}
		col := validCols[rand.Intn(len(validCols))]
		row := getAvailableRow(board, col)
		if row != -1 {
			dropPiece(board, row, col, currentPlayer)
			if currentPlayer == P1 {
				currentPlayer = P2
			} else {
				currentPlayer = P1
			}
		}
		moves--
	}

	return board
}

func printBoard(board Board) {
	for _, row := range board {
		for _, cell := range row {
			fmt.Printf("%d ", cell)
		}
		fmt.Println()
	}
}

func main() {
	board := createBoard()
	printBoard(simulateGame(board, 0))

	start := time.Now()
	tree := exploreGameTree(board, P1, 8)
	elapsed := time.Since(start)

	fmt.Printf("\n⏱️ Temps d'exécution de exploreGameTree : %s\n", elapsed)

	path := findShortestWinningPath(tree, WIN_P1)

	fmt.Println("\n✅ Chemin gagnant le plus logique et le plus court pour P1 (avec coups P2 inclus) :")
	if path != nil {
		for i, move := range path {
			turn := "P1"
			if i%2 != 0 {
				turn = "P2"
			}
			fmt.Printf("%s joue colonne %d\n", turn, move)
		}
	} else {
		fmt.Println("Aucun chemin gagnant trouvé.")
	}
}
