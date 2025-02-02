package main

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Loc [4][2]int

type Tile uint

const (
	LightBlue Tile = iota
	Yellow
	Purple
	Orange
	Blue
	Green
	Red
	Black
	Ghost
)

var styleMap = map[Tile]tcell.Style{
	LightBlue: tcell.StyleDefault.Background(tcell.ColorAqua),
	Yellow:    tcell.StyleDefault.Background(tcell.ColorYellow),
	Purple:    tcell.StyleDefault.Background(tcell.ColorFuchsia),
	Orange:    tcell.StyleDefault.Background(tcell.ColorCoral),
	Blue:      tcell.StyleDefault.Background(tcell.ColorNavy),
	Green:     tcell.StyleDefault.Background(tcell.ColorLime),
	Red:       tcell.StyleDefault.Background(tcell.Color160),
	Black:     tcell.StyleDefault.Background(tcell.ColorBlack),
	Ghost:     tcell.StyleDefault.Background(tcell.ColorGray),
}

var initLoc = map[Tile]Loc{
	LightBlue: {
		{-1, 3},
		{-1, 4},
		{-1, 5},
		{-1, 6},
	},
	Yellow: {
		{-1, 4},
		{-1, 5},
		{-2, 4},
		{-2, 5},
	},
	Purple: {
		{-1, 4},
		{-1, 5},
		{-1, 6},
		{-2, 5},
	},
	Orange: {
		{-1, 4},
		{-1, 5},
		{-1, 6},
		{-2, 6},
	},
	Blue: {
		{-1, 4},
		{-1, 5},
		{-1, 6},
		{-2, 4},
	},
	Green: {
		{-1, 4},
		{-1, 5},
		{-2, 5},
		{-2, 6},
	},
	Red: {
		{-1, 5},
		{-1, 6},
		{-2, 4},
		{-2, 5},
	},
}

var wallKick3 = map[int][][]int{
	0: {
		{0, -1},
		{-1, -1},
		{2, 0},
		{2, -1},
	},
	1: {
		{0, 1},
		{1, 1},
		{-2, 0},
		{-2, 1},
	},
	2: {
		{0, 1},
		{-1, 1},
		{2, 0},
		{2, 1},
	},
	3: {
		{0, -1},
		{1, -1},
		{-2, 0},
		{-2, -1},
	},
}

var wallKick4 = map[int][][]int{
	0: {
		{0, -2},
		{0, 1},
		{1, -2},
		{-2, 1},
	},
	1: {
		{0, -1},
		{0, 2},
		{-2, -1},
		{1, 2},
	},
	2: {
		{0, 2},
		{0, -1},
		{1, -2},
		{-2, 1},
	},
	3: {
		{0, 1},
		{0, -2},
		{2, 1},
		{-1, -2},
	},
}

var rotation = map[Tile]map[int]Loc{
	LightBlue: {
		0: {
			{-1, 2},
			{0, 1},
			{1, 0},
			{2, -1},
		},
		1: {
			{2, 1},
			{1, 0},
			{0, -1},
			{-1, -2},
		},
		2: {
			{1, -2},
			{0, -1},
			{-1, 0},
			{-2, 1},
		},
		3: {
			{-2, -1},
			{-1, 0},
			{0, 1},
			{1, 2},
		},
	},
	Purple: {
		0: {
			{-1, 1},
			{0, 0},
			{1, -1},
			{1, 1},
		},
		1: {
			{1, 1},
			{0, 0},
			{-1, -1},
			{1, -1},
		},
		2: {
			{1, -1},
			{0, 0},
			{-1, 1},
			{-1, -1},
		},
		3: {
			{-1, -1},
			{0, 0},
			{1, 1},
			{-1, 1},
		},
	},
	Orange: {
		0: {
			{-1, 1},
			{0, 0},
			{1, -1},
			{2, 0},
		},
		1: {
			{1, 1},
			{0, 0},
			{-1, -1},
			{0, -2},
		},
		2: {
			{1, -1},
			{0, 0},
			{-1, 1},
			{-2, 0},
		},
		3: {
			{-1, -1},
			{0, 0},
			{1, 1},
			{0, 2},
		},
	},
	Blue: {
		0: {
			{-1, 1},
			{0, 0},
			{1, -1},
			{0, 2},
		},
		1: {
			{1, 1},
			{0, 0},
			{-1, -1},
			{2, 0},
		},
		2: {
			{1, -1},
			{0, 0},
			{-1, 1},
			{0, -2},
		},
		3: {
			{-1, -1},
			{0, 0},
			{1, 1},
			{-2, 0},
		},
	},
	Green: {
		0: {
			{-1, 1},
			{0, 0},
			{1, 1},
			{2, 0},
		},
		1: {
			{1, 1},
			{0, 0},
			{1, -1},
			{0, -2},
		},
		2: {
			{1, -1},
			{0, 0},
			{-1, -1},
			{-2, 0},
		},
		3: {
			{-1, -1},
			{0, 0},
			{-1, 1},
			{0, 2},
		},
	},
	Red: {
		0: {
			{0, 0},
			{1, -1},
			{0, 2},
			{1, 1},
		},
		1: {
			{0, 0},
			{-1, -1},
			{2, 0},
			{1, -1},
		},
		2: {
			{0, 0},
			{-1, 1},
			{0, -2},
			{-1, -1},
		},
		3: {
			{0, 0},
			{1, 1},
			{-2, 0},
			{-1, 1},
		},
	},
}

func (t Tile) GetStyle() tcell.Style {
	return styleMap[t]
}

func NewPiece() *Piece {
	rand.Seed(time.Now().UTC().UnixNano())
	tile := Tile(rand.Intn(7))
	piece := &Piece{
		tile,
		initLoc[tile],
		0,
	}

	return piece
}

type Piece struct {
	tile     Tile
	loc      Loc
	rotation int
}

func (p *Piece) Reset() {
	p.loc = initLoc[p.tile]
	p.rotation = 0
}

func (p *Piece) Down() Loc {
	newLoc := Loc{}

	for i, loc := range p.loc {
		newLoc[i] = [2]int{loc[0] + 1, loc[1]}
	}

	return newLoc
}

func (p *Piece) Left() Loc {
	newLoc := Loc{}

	for i, loc := range p.loc {
		newLoc[i] = [2]int{loc[0], loc[1] - 1}
	}

	return newLoc
}

func (p *Piece) Right() Loc {
	newLoc := Loc{}

	for i, loc := range p.loc {
		newLoc[i] = [2]int{loc[0], loc[1] + 1}
	}

	return newLoc
}

func (p *Piece) Rotate() Loc {
	newLoc := Loc{}
	t := rotation[p.tile][p.rotation]
	for i, loc := range p.loc {
		newLoc[i] = [2]int{loc[0] + t[i][0], loc[1] + t[i][1]}
	}

	return newLoc
}

func (p *Piece) Clear(board [][]Tile) {
	for _, loc := range p.loc {
		row, col := loc[0], loc[1]
		if row >= 0 {
			board[row][col] = Black
		}
	}
}

func (p *Piece) Render(board [][]Tile) {
	for _, loc := range p.loc {
		row, col := loc[0], loc[1]
		if row >= 0 {
			board[row][col] = p.tile
		}
	}
}
