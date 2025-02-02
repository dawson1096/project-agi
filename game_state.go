package main

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

func NewGameState() *GameState {
	g := &GameState{}
	board := make([][]Tile, 20)
	for i := range board {
		board[i] = make([]Tile, 10)
	}
	g.board = board

	g.Reset()

	return g
}

type GameState struct {
	board     [][]Tile
	score     int
	canHold   bool
	hold      *Piece
	next      *Piece
	curr      *Piece
	lastDrop  time.Time
	toLock    time.Time
	needLock  bool
	clearTime time.Time
	toClear   []int
	level     int
	clears    int
	started   bool
	pause     bool
	gameover  bool

	sync.Mutex
}

func (g *GameState) Start() {
	g.Reset()
	g.started = true
	g.gameover = false
	g.lastDrop = time.Now()
	g.curr = NewPiece()
	g.next = NewPiece()
}

func (g *GameState) Update(t time.Time) {
	g.Lock()
	defer g.Unlock()

	if !g.started || g.pause {
		return
	}

	if len(g.toClear) > 0 {
		clearDelay := time.Duration(0.1 * float64(time.Second))
		if time.Now().After(g.clearTime.Add(clearDelay)) {
			g.clearLines()
			g.spawnNext()
		}
		return
	}

	if g.needLock {
		lockDelay := time.Duration(0.5 * float64(time.Second))
		if time.Now().After(g.toLock.Add(lockDelay)) {
			if g.tryMove(g.curr.Down()) {
				g.lastDrop = time.Now()
			} else {
				g.placePiece()
			}
			g.needLock = false
		}
		return
	}

	durFloat := math.Pow((0.8 - ((float64(g.level) - 1) * 0.007)), float64(g.level-1))
	dur := time.Duration(durFloat * float64(time.Second))
	if time.Now().After(g.lastDrop.Add(dur)) {
		if !g.tryMove(g.curr.Down()) {
			g.toLock = time.Now()
			g.needLock = true
		}
		g.lastDrop = time.Now()
	}

}

func (g *GameState) HandleEvent(ev *tcell.EventKey) {
	g.Lock()
	defer g.Unlock()

	if g.pause && ev.Rune() != 'p' {
		return
	}

	if ev.Rune() == 'p' {
		g.pause = !g.pause
	} else if ev.Rune() == 'h' || ev.Key() == tcell.KeyLeft {
		g.tryMove(g.curr.Left())
	} else if ev.Rune() == 'l' || ev.Key() == tcell.KeyRight {
		g.tryMove(g.curr.Right())
	} else if ev.Rune() == 'j' || ev.Key() == tcell.KeyDown {
		if !g.tryMove(g.curr.Down()) {
			g.placePiece()
		}
		g.lastDrop = time.Now()
	} else if ev.Rune() == 'k' || ev.Key() == tcell.KeyUp {
		g.tryRotate()
	} else if ev.Rune() == ' ' {
		for g.tryMove(g.curr.Down()) {
		}
		g.placePiece()
		g.lastDrop = time.Now()
	} else if ev.Rune() == 'f' {
		g.holdPiece()
	}
}

func (g *GameState) Reset() {
	for i, row := range g.board {
		for j := range row {
			g.board[i][j] = Black
		}
	}
	g.score = 0
	g.canHold = true
	g.hold = nil
	g.next = nil
	g.curr = nil
	g.level = 1
	g.clears = 0
	g.started = false
	g.pause = false
	g.gameover = false
}

func (g *GameState) tryMove(newLoc Loc) bool {
	g.curr.Clear(g.board)

	valid := true
	for _, loc := range newLoc {
		row, col := loc[0], loc[1]
		if row > 19 || col < 0 || col > 9 {
			valid = false
			newLoc = g.curr.loc
			break
		}

		if row >= 0 && g.board[row][col] != Black {
			valid = false
			newLoc = g.curr.loc
			break
		}
	}

	g.curr.loc = newLoc
	g.curr.Render(g.board)

	return valid
}

func (g *GameState) tryRotate() bool {
	if g.curr.tile == Yellow {
		return true
	}

	newLoc := g.curr.Rotate()
	if g.tryMove(newLoc) {
		g.curr.rotation = (g.curr.rotation + 1) % 4
		return true
	}

	var wallKick [][]int
	if g.curr.tile == LightBlue {
		wallKick = wallKick4[g.curr.rotation]
	} else {
		wallKick = wallKick3[g.curr.rotation]
	}

	for _, t := range wallKick {
		kickLoc := Loc{}
		for i, loc := range newLoc {
			kickLoc[i] = [2]int{loc[0] + t[0], loc[1] + t[1]}
		}
		if g.tryMove(kickLoc) {
			g.curr.rotation = (g.curr.rotation + 1) % 4
			return true
		}
	}
	return false
}

func (g *GameState) placePiece() {
	for _, loc := range g.curr.loc {
		if loc[0] < 0 {
			g.gameover = true
			g.started = false
			return
		}
	}

	if !g.checkClears() {
		g.spawnNext()
	}
}

func (g *GameState) checkClears() bool {
	unique := make(map[int]bool)
	clear := []int{}
	for _, loc := range g.curr.loc {
		if !unique[loc[0]] {
			unique[loc[0]] = true
			hole := false
			for _, tile := range g.board[loc[0]] {
				if tile == Black {
					hole = true
					break
				}
			}
			if !hole {
				clear = append(clear, loc[0])
			}
		}
	}
	sort.Slice(clear, func(i, j int) bool {
		return clear[i] > clear[j]
	})

	switch len(clear) {
	case 1:
		g.score += 40 * (g.level + 1)
	case 2:
		g.score += 100 * (g.level + 1)
	case 3:
		g.score += 300 * (g.level + 1)
	case 4:
		g.score += 1200 * (g.level + 1)
	}
	g.clears += len(clear)
	if g.clears >= 10 {
		g.level++
		g.clears = g.clears - 10
	}

	g.toClear = clear
	g.clearTime = time.Now()

	return len(clear) > 0
}

func (g *GameState) clearLines() {
	for _, row := range g.toClear {
		for col := range g.board[row] {
			g.board[row][col] = Black
		}
	}
	shift := 0
	for len(g.toClear) > 0 {
		row := g.toClear[0] + shift
		shift++
		g.toClear = g.toClear[1:]

		for i := row; i > 0; i-- {
			for j := 0; j < 10; j++ {
				g.board[i][j] = g.board[i-1][j]
			}
		}
		for j := 0; j < 10; j++ {
			g.board[0][j] = Black
		}
	}
}

func (g *GameState) spawnNext() {
	g.curr = g.next
	g.next = NewPiece()
	g.canHold = true
}

func (g *GameState) holdPiece() {
	if g.canHold {
		g.curr.Clear(g.board)
		if g.hold == nil {
			g.hold = g.curr
			g.hold.Reset()
			g.curr = g.next
			g.next = NewPiece()
		} else {
			g.next = g.curr
			g.next.Reset()
			g.curr = g.hold
			g.hold = nil
			g.canHold = false
		}
	}
}
