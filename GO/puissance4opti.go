package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Player int
type Cell int

const (
	Empty Cell   = 0
	P1    Player = 1
	P2    Player = 2

	Rows = 6
	Cols = 7
)

type Board [Rows][Cols]Cell

func createBoard() Board {
	return Board{}
}

func getAvailableRow(b *Board, col int) int {
	for row := Rows - 1; row >= 0; row-- {
		if b[row][col] == Empty {
			return row
		}
	}
	return -1
}

func dropPiece(b *Board, row, col int, player Player) {
	b[row][col] = Cell(player)
}

func isWinningMove(b *Board, player Player) bool {
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}
	p := Cell(player)

	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b[r][c] != p {
				continue
			}
			for _, d := range directions {
				count := 0
				nr, nc := r, c
				for i := 0; i < 4; i++ {
					if nr < 0 || nr >= Rows || nc < 0 || nc >= Cols || b[nr][nc] != p {
						break
					}
					count++
					nr += d[0]
					nc += d[1]
				}
				if count == 4 {
					return true
				}
			}
		}
	}
	return false
}

func getValidColumns(b *Board) []int {
	cols := make([]int, 0, Cols)
	for c := 0; c < Cols; c++ {
		if b[0][c] == Empty {
			cols = append(cols, c)
		}
	}
	return cols
}

type GameOutcome string

const (
	WinP1 GameOutcome = "WIN_P1"
	WinP2 GameOutcome = "WIN_P2"
	Draw  GameOutcome = "DRAW"
)

type GameNode struct {
	Move     int
	Outcome  GameOutcome
	Children []GameNode
}

func summarizeOutcomes(nodes []GameNode, player Player) GameOutcome {
	win := WinP1
	lose := WinP2
	if player == P2 {
		win, lose = WinP2, WinP1
	}

	for _, n := range nodes {
		if n.Outcome == win {
			return win
		}
	}
	allLose := true
	for _, n := range nodes {
		if n.Outcome != lose {
			allLose = false
			break
		}
	}
	if allLose {
		return lose
	}
	return Draw
}

func outcomeFromPlayer(player Player) GameOutcome {
	if player == P1 {
		return WinP1
	}
	return WinP2
}

func exploreGameTreeSmart(board Board, currentPlayer Player, depth int) []GameNode {
	if depth == 0 {
		return []GameNode{{Move: -1, Outcome: Draw}}
	}

	validCols := getValidColumns(&board)
	if len(validCols) == 0 {
		return []GameNode{{Move: -1, Outcome: Draw}}
	}

	opponent := P1
	if currentPlayer == P1 {
		opponent = P2
	}

	type Task struct {
		Index int
		Col   int
		Board Board
	}

	type Result struct {
		Index int
		Node  GameNode
	}

	tasks := make(chan Task, len(validCols))
	results := make(chan Result, len(validCols))

	const numWorkers = 8

	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range tasks {
				row := getAvailableRow(&task.Board, task.Col)
				if row == -1 {
					continue
				}
				newBoard := task.Board
				dropPiece(&newBoard, row, task.Col, currentPlayer)

				if isWinningMove(&newBoard, currentPlayer) {
					results <- Result{
						Index: task.Index,
						Node:  GameNode{Move: task.Col, Outcome: outcomeFromPlayer(currentPlayer)},
					}
				} else {
					children := exploreGameTreeSmart(newBoard, opponent, depth-1)
					outcome := summarizeOutcomes(children, currentPlayer)
					results <- Result{
						Index: task.Index,
						Node:  GameNode{Move: task.Col, Outcome: outcome, Children: children},
					}
				}
			}
		}()
	}

	for i, col := range validCols {
		tasks <- Task{Index: i, Col: col, Board: board}
	}
	close(tasks)

	resultNodes := make([]GameNode, len(validCols))
	for i := 0; i < len(validCols); i++ {
		res := <-results
		resultNodes[res.Index] = res.Node
	}

	return resultNodes
}

func findShortestWinningPath(nodes []GameNode, target GameOutcome) []int {
	var shortest []int
	stack := []struct {
		path     []int
		children []GameNode
		player   Player
	}{{
		path:     []int{},
		children: nodes,
		player:   P1,
	}}

	for len(stack) > 0 {
		n := len(stack) - 1
		item := stack[n]
		stack = stack[:n]

		for _, node := range item.children {
			newPath := append(item.path, node.Move)
			if node.Outcome == target && len(node.Children) == 0 {
				if shortest == nil || len(newPath) < len(shortest) {
					shortest = append([]int(nil), newPath...)
				}
			} else if len(node.Children) > 0 {
				stack = append(stack, struct {
					path     []int
					children []GameNode
					player   Player
				}{
					path:     newPath,
					children: node.Children,
					player:   3 - item.player,
				})
			}
		}
	}
	return shortest
}

func simulateGame(board *Board, moves int) {
	currentPlayer := P1
	for moves > 0 {
		validCols := getValidColumns(board)
		if len(validCols) == 0 {
			break
		}
		col := validCols[rand.Intn(len(validCols))]
		row := getAvailableRow(board, col)
		if row != -1 {
			dropPiece(board, row, col, currentPlayer)
			currentPlayer = 3 - currentPlayer
		}
		moves--
	}
}

func printBoard(b Board) {
	for _, row := range b {
		fmt.Println(row)
	}
}

func main() {
	start := time.Now()

	rand.Seed(time.Now().UnixNano())

	board := createBoard()
	simulateGame(&board, 0)

	fmt.Println("Initial board:")
	printBoard(board)

	tree := exploreGameTreeSmart(board, P1, 8)
	path := findShortestWinningPath(tree, WinP1)

	fmt.Println("\n✅ Chemin gagnant le plus logique et le plus court pour P1 (avec coups P2 inclus) :")
	if path != nil {
		for i, move := range path {
			player := "P1"
			if i%2 == 1 {
				player = "P2"
			}
			fmt.Printf("%s joue colonne %d\n", player, move)
		}
	} else {
		fmt.Println("Aucun chemin gagnant trouvé.")
	}

	elapsed := time.Since(start)
	fmt.Printf("\n⏱️ Temps d'exécution : %s\n", elapsed)
}
