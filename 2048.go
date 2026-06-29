// 2048.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	red     = "\033[91m"
	green   = "\033[92m"
	yellow  = "\033[93m"
	blue    = "\033[94m"
	magenta = "\033[95m"
	cyan    = "\033[96m"
	white   = "\033[97m"
	bgRed   = "\033[101m"
	bgGreen = "\033[102m"
	bgYellow= "\033[103m"
	bgBlue  = "\033[104m"
	bgMagenta="\033[105m"
	bgCyan  = "\033[106m"
	bgWhite = "\033[107m"
)

func colorize(text, color string) string {
	return color + text + reset
}

type TileColors struct {
	bg string
	fg string
}

func getTileColor(val int) TileColors {
	switch val {
	case 2: return TileColors{bgYellow, "black"}
	case 4: return TileColors{bgBlue, white}
	case 8: return TileColors{bgCyan, "black"}
	case 16: return TileColors{bgGreen, white}
	case 32: return TileColors{bgMagenta, white}
	case 64: return TileColors{bgRed, white}
	case 128: return TileColors{bgYellow, "black"}
	case 256: return TileColors{bgBlue, white}
	case 512: return TileColors{bgCyan, "black"}
	case 1024: return TileColors{bgMagenta, white}
	case 2048: return TileColors{bgRed, white}
	case 4096: return TileColors{bgGreen, "black"}
	case 8192: return TileColors{bgBlue, white}
	default: return TileColors{bgWhite, "black"}
	}
}

type RecordData map[string]int

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func getRecordFile() string {
	return getHomeDir() + "/.2048_records.json"
}

func loadRecords() RecordData {
	var rec RecordData
	data, err := os.ReadFile(getRecordFile())
	if err != nil {
		return make(RecordData)
	}
	json.Unmarshal(data, &rec)
	return rec
}

func saveRecords(rec RecordData) {
	data, _ := json.MarshalIndent(rec, "", "  ")
	os.WriteFile(getRecordFile(), data, 0644)
}

type Game2048 struct {
	size    int
	board   [][]int
	score   int
	best    int
	history []struct {
		board [][]int
		score int
	}
	records RecordData
}

func NewGame(size int) *Game2048 {
	g := &Game2048{
		size:    size,
		board:   make([][]int, size),
		score:   0,
		history: []struct {
			board [][]int
			score int
		}{},
	}
	for i := range g.board {
		g.board[i] = make([]int, size)
	}
	g.records = loadRecords()
	g.best = g.records[strconv.Itoa(size)]
	return g
}

func (g *Game2048) addRandomTile() {
	var empty [][2]int
	for i := 0; i < g.size; i++ {
		for j := 0; j < g.size; j++ {
			if g.board[i][j] == 0 {
				empty = append(empty, [2]int{i, j})
			}
		}
	}
	if len(empty) == 0 {
		return
	}
	idx := rand.Intn(len(empty))
	i, j := empty[idx][0], empty[idx][1]
	if rand.Float64() < 0.9 {
		g.board[i][j] = 2
	} else {
		g.board[i][j] = 4
	}
}

func (g *Game2048) compress(row []int) []int {
	res := []int{}
	for _, x := range row {
		if x != 0 {
			res = append(res, x)
		}
	}
	for len(res) < g.size {
		res = append(res, 0)
	}
	return res
}

func (g *Game2048) merge(row []int) ([]int, int) {
	res := []int{}
	add := 0
	i := 0
	for i < len(row) {
		if i < len(row)-1 && row[i] == row[i+1] && row[i] != 0 {
			res = append(res, row[i]*2)
			add += row[i] * 2
			i += 2
		} else {
			res = append(res, row[i])
			i++
		}
	}
	for len(res) < g.size {
		res = append(res, 0)
	}
	return res, add
}

func (g *Game2048) move(dir string) bool {
	prevBoard := make([][]int, g.size)
	for i := range prevBoard {
		prevBoard[i] = make([]int, g.size)
		copy(prevBoard[i], g.board[i])
	}
	prevScore := g.score
	addScore := 0

	switch dir {
	case "w":
		for j := 0; j < g.size; j++ {
			col := []int{}
			for i := 0; i < g.size; i++ {
				col = append(col, g.board[i][j])
			}
			col = g.compress(col)
			merged, add := g.merge(col)
			merged = g.compress(merged)
			for i := 0; i < g.size; i++ {
				g.board[i][j] = merged[i]
			}
			addScore += add
		}
	case "s":
		for j := 0; j < g.size; j++ {
			col := []int{}
			for i := g.size - 1; i >= 0; i-- {
				col = append(col, g.board[i][j])
			}
			col = g.compress(col)
			merged, add := g.merge(col)
			merged = g.compress(merged)
			for i := 0; i < g.size; i++ {
				g.board[g.size-1-i][j] = merged[i]
			}
			addScore += add
		}
	case "a":
		for i := 0; i < g.size; i++ {
			row := g.compress(g.board[i])
			merged, add := g.merge(row)
			merged = g.compress(merged)
			g.board[i] = merged
			addScore += add
		}
	case "d":
		for i := 0; i < g.size; i++ {
			row := make([]int, g.size)
			copy(row, g.board[i])
			reverse(row)
			row = g.compress(row)
			merged, add := g.merge(row)
			merged = g.compress(merged)
			reverse(merged)
			g.board[i] = merged
			addScore += add
		}
	}

	if !equalBoard(g.board, prevBoard) {
		g.history = append(g.history, struct {
			board [][]int
			score int
		}{prevBoard, prevScore})
		g.score += addScore
		if g.score > g.best {
			g.best = g.score
			key := strconv.Itoa(g.size)
			g.records[key] = g.best
			saveRecords(g.records)
		}
		g.addRandomTile()
		return true
	}
	return false
}

func (g *Game2048) undo() bool {
	if len(g.history) == 0 {
		return false
	}
	last := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.board = last.board
	g.score = last.score
	return true
}

func (g *Game2048) isWin() bool {
	for _, row := range g.board {
		for _, v := range row {
			if v >= 2048 {
				return true
			}
		}
	}
	return false
}

func (g *Game2048) isGameOver() bool {
	for i := 0; i < g.size; i++ {
		for j := 0; j < g.size; j++ {
			if g.board[i][j] == 0 {
				return false
			}
			if i < g.size-1 && g.board[i][j] == g.board[i+1][j] {
				return false
			}
			if j < g.size-1 && g.board[i][j] == g.board[i][j+1] {
				return false
			}
		}
	}
	return true
}

func (g *Game2048) display() {
	clearScreen()
	fmt.Println(colorize(fmt.Sprintf("🎮  2048  |  Размер %d×%d  |  Счёт: %d  |  Лучший: %d",
		g.size, g.size, g.score, g.best), bold))
	fmt.Println(strings.Repeat("─", g.size*8+1))
	for i := 0; i < g.size; i++ {
		for j := 0; j < g.size; j++ {
			val := g.board[i][j]
			if val == 0 {
				fmt.Print(colorize("      ", bgWhite) + " ")
			} else {
				tc := getTileColor(val)
				text := fmt.Sprintf("%6d", val)
				fmt.Print(colorize(text, tc.bg) + " ")
			}
			if j < g.size-1 {
				fmt.Print("│")
			}
		}
		fmt.Println()
		if i < g.size-1 {
			fmt.Println(strings.Repeat("─", g.size*8+1))
		}
	}
	fmt.Println(strings.Repeat("─", g.size*8+1))
	fmt.Println("WASD — движение | U — отмена | Q — выход")
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func equalBoard(a, b [][]int) bool {
	for i := range a {
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	size := 4
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-s" && i+1 < len(os.Args) {
			size, _ = strconv.Atoi(os.Args[i+1])
			i++
		} else if os.Args[i] == "-h" || os.Args[i] == "--help" {
			fmt.Println("Usage: 2048 [options]\n  -s <N>   Size (4, 5, 6) default 4")
			return
		}
	}
	if size < 4 || size > 6 {
		size = 4
	}

	rand.Seed(time.Now().UnixNano())
	game := NewGame(size)
	game.addRandomTile()
	game.addRandomTile()

	reader := bufio.NewReader(os.Stdin)
	for {
		game.display()
		if game.isWin() {
			fmt.Println(colorize("🎉 Поздравляем! Вы достигли 2048!", bold))
		}
		if game.isGameOver() {
			fmt.Println(colorize("💀 Игра окончена. Нажмите Q для выхода.", red))
		}

		char, _ := reader.ReadString('\n')
		char = strings.TrimSpace(strings.ToLower(char))
		if char == "q" {
			fmt.Println(colorize("Игра сохранена. До встречи!", yellow))
			break
		} else if char == "u" {
			if game.undo() {
				fmt.Println(colorize("Ход отменён.", yellow))
			} else {
				fmt.Println(colorize("Нечего отменять.", yellow))
			}
		} else if char == "w" {
			game.move("w")
		} else if char == "a" {
			game.move("a")
		} else if char == "s" {
			game.move("s")
		} else if char == "d" {
			game.move("d")
		}
	}
}
