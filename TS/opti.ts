type Player = 1 | 2;
type Cell = 0 | Player;
type Board = Cell[][];

const ROWS = 6;
const COLS = 7;

function createBoard(): Board {
    return Array.from({ length: ROWS }, () => Array(COLS).fill(0));
}

function cloneBoard(board: Board): Board {
    return board.map(row => [...row]);
}

function getAvailableRow(board: Board, col: number): number | null {
    for (let row = ROWS - 1; row >= 0; row--) {
        if (board[row][col] === 0) return row;
    }
    return null;
}

function dropPiece(board: Board, row: number, col: number, player: Player): void {
    board[row][col] = player;
}

function isWinningMove(board: Board, player: Player): boolean {
    const directions = [
        [0, 1], [1, 0], [1, 1], [1, -1]
    ];

    for (let r = 0; r < ROWS; r++) {
        for (let c = 0; c < COLS; c++) {
            if (board[r][c] !== player) continue;
            for (const [dr, dc] of directions) {
                let count = 0;
                for (let i = 0; i < 4; i++) {
                    const nr = r + dr * i;
                    const nc = c + dc * i;
                    if (nr >= 0 && nr < ROWS && nc >= 0 && nc < COLS && board[nr][nc] === player) {
                        count++;
                    }
                }
                if (count === 4) return true;
            }
        }
    }
    return false;
}

function getValidColumns(board: Board): number[] {
    return [...Array(COLS).keys()].filter(col => getAvailableRow(board, col) !== null);
}

type GameOutcome = "WIN_P1" | "WIN_P2" | "DRAW";
type GameNode = {
    move: number;
    outcome: GameOutcome;
    children: GameNode[];
};

async function exploreGameTreeSmart(board: Board, currentPlayer: Player, depth: number): Promise<GameNode[]> {
    const opponent = currentPlayer === 1 ? 2 : 1;
    const validCols = getValidColumns(board).sort((a, b) => Math.abs(3 - a) - Math.abs(3 - b));

    if (depth === 0 || validCols.length === 0) {
        return [{ move: -1, outcome: "DRAW", children: [] }];
    }

    const results: GameNode[] = [];

    for (const col of validCols) {
        const row = getAvailableRow(board, col);
        if (row === null) continue;

        const newBoard = cloneBoard(board);
        dropPiece(newBoard, row, col, currentPlayer);

        if (isWinningMove(newBoard, currentPlayer)) {
            return [{
                move: col,
                outcome: currentPlayer === 1 ? "WIN_P1" : "WIN_P2",
                children: [],
            }];
        }

        let children: GameNode[];

        if (depth <= 3) {
            children = await Promise.all(
                getValidColumns(newBoard).map(async subCol => {
                    const subRow = getAvailableRow(newBoard, subCol);
                    if (subRow === null) return null;

                    const subBoard = cloneBoard(newBoard);
                    dropPiece(subBoard, subRow, subCol, opponent);

                    if (isWinningMove(subBoard, opponent)) {
                        return {
                            move: subCol,
                            outcome: opponent === 1 ? "WIN_P1" : "WIN_P2",
                            children: [],
                        };
                    }

                    return {
                        move: subCol,
                        outcome: "DRAW", 
                        children: [],
                    };
                })
            ).then(res => res.filter(Boolean) as GameNode[]);
        } else {
            children = await exploreGameTreeSmart(newBoard, opponent, depth - 1);
        }

        const outcome = summarizeOutcomes(children, currentPlayer);
        const result = { move: col, outcome, children };

        if (outcome === (currentPlayer === 1 ? "WIN_P1" : "WIN_P2")) {
            return [result];
        }

        results.push(result);
    }

    return results;
}

function summarizeOutcomes(nodes: GameNode[], player: Player): GameOutcome {
    const win = player === 1 ? "WIN_P1" : "WIN_P2";
    const lose = player === 1 ? "WIN_P2" : "WIN_P1";

    if (nodes.some(n => n.outcome === win)) return win;
    if (nodes.every(n => n.outcome === lose)) return lose;
    return "DRAW";
}

function findShortestWinningPath(nodes: GameNode[], target: GameOutcome): number[] | null {
    let shortest: number[] | null = null;

    function dfs(path: number[], children: GameNode[], depth: number, currentPlayer: Player) {
        for (const node of children) {
            const newPath = [...path, node.move];
            if (node.outcome === target && node.children.length === 0) {
                if (!shortest || newPath.length < shortest.length) {
                    shortest = newPath;
                }
            } else if (node.children.length > 0) {
                dfs(newPath, node.children, depth + 1, currentPlayer === 1 ? 2 : 1);
            }
        }
    }

    dfs([], nodes, 0, 1);
    return shortest;
}

function simulateGame(board: Board, moves: number): Board {
    let currentPlayer: Player = 1;

    while (moves-- > 0) {
        const validCols = getValidColumns(board);
        if (validCols.length === 0) break;

        const col = validCols[Math.floor(Math.random() * validCols.length)];
        const row = getAvailableRow(board, col);
        if (row !== null) {
            dropPiece(board, row, col, currentPlayer);
            currentPlayer = currentPlayer === 1 ? 2 : 1;
        }
    }

    return board;
}

// 🧪 Test
const board = createBoard();
console.table(
    simulateGame(board, 6)
);

(async () => {
    const tree = await exploreGameTreeSmart(board, 1, 8);

    const path = findShortestWinningPath(tree, "WIN_P1");

    console.log("\n✅ Chemin gagnant le plus logique et le plus court pour P1 (avec coups P2 inclus) :");
    if (path) {
        path.forEach((move, index) => {
            const turn = index % 2 === 0 ? "P1" : "P2";
            console.log(`${turn} joue colonne ${move}`);
        });
    } else {
        console.log("Aucun chemin gagnant trouvé.");
    }
})();
