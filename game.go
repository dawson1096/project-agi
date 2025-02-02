package main

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type Game struct {
	quitq     chan struct{}
	screen    tcell.Screen
	eventq    chan tcell.Event
	errmsg    string
	quitone   sync.Once
	gameState *GameState
}

func (g *Game) Init() error {
	if screen, err := tcell.NewScreen(); err != nil {
		return err
	} else if err = screen.Init(); err != nil {
		return err
	} else {
		screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack))
		g.screen = screen
	}

	g.gameState = NewGameState()
	g.quitq = make(chan struct{})
	g.eventq = make(chan tcell.Event)

	return nil
}

func (g *Game) Quit() {
	g.quitone.Do(func() {
		close(g.quitq)
	})
}

func (g *Game) Draw() {
	g.gameState.Lock()
	width, height := g.screen.Size()
	g.drawBoard(width, height)
	g.drawNextBox(width, height)
	g.drawHoldBox(width, height)
	g.drawScore(width, height)
	g.drawHelp(width, height)
	g.drawMessage(width, height)
	g.screen.Show()
	g.gameState.Unlock()
}

func (g *Game) Run() error {
	go g.EventPoller()
	go g.Updater()
loop:
	for {
		g.Draw()
		select {
		case <-g.quitq:
			break loop
		case <-time.After(time.Millisecond * 10):
		case ev := <-g.eventq:
			g.HandleEvent(ev)
		}
	}

	// Inject a wakeup interrupt
	iev := tcell.NewEventInterrupt(nil)
	g.screen.PostEvent(iev)

	g.screen.Fini()
	// wait for updaters to finish
	if g.errmsg != "" {
		return errors.New(g.errmsg)
	}
	return nil
}

func (g *Game) Error(msg string) {
	g.errmsg = msg
	g.Quit()
}

func (g *Game) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventResize:
		g.screen.Clear()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			g.Quit()
			return true
		}
		if !g.gameState.started || g.gameState.gameover {
			if ev.Key() == tcell.KeyEnter {
				g.gameState.Start()
			}
		} else {
			g.gameState.HandleEvent(ev)
		}
	}

	return true
}

func (g *Game) Updater() {
	for {
		select {
		case <-g.quitq:
			return
		case <-time.After(time.Millisecond * 10):
			g.gameState.Update(time.Now())
		}
	}
}

func (g *Game) EventPoller() {
	for {
		ev := g.screen.PollEvent()
		if ev == nil {
			return
		}
		select {
		case <-g.quitq:
			return
		case g.eventq <- ev:
		}
	}
}

func (g *Game) drawBoard(width, height int) {
	x := width/2 - 10
	y := height/2 - 10
	for i, row := range g.gameState.board {
		for j, tile := range row {
			style := tile.GetStyle()
			g.screen.SetContent(x+(j*2), y+i, ' ', nil, style)
			g.screen.SetContent(x+(j*2)+1, y+i, ' ', nil, style)
		}
	}
}

func (g *Game) drawNextBox(width, height int) {
	x := width/2 + 10
	y := height/2 - 10

	boxStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	headStyle := tcell.StyleDefault.Background(tcell.ColorGray)
	for col := x + 1; col < x+15; col++ {
		g.screen.SetContent(col, y, ' ', nil, headStyle)
	}
	col := x + 6
	for _, r := range "Next" {
		g.screen.SetContent(col, y, r, nil, headStyle)
		col++
	}
	for row := y + 1; row < y+7; row++ {
		for col := x + 1; col < x+15; col++ {
			g.screen.SetContent(col, row, ' ', nil, boxStyle)
		}
	}

	if g.gameState.next != nil {
		g.drawPiece(g.gameState.next, x+1, y)
	}
}

func (g *Game) drawHoldBox(width, height int) {
	x := width/2 - 10
	y := height/2 - 10

	boxStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	headStyle := tcell.StyleDefault.Background(tcell.ColorGray)
	for col := x - 15; col < x-1; col++ {
		g.screen.SetContent(col, y, ' ', nil, headStyle)
	}
	col := x - 10
	for _, r := range "Hold" {
		g.screen.SetContent(col, y, r, nil, headStyle)
		col++
	}
	for row := y + 1; row < y+7; row++ {
		for col := x - 15; col < x-1; col++ {
			g.screen.SetContent(col, row, ' ', nil, boxStyle)
		}
	}

	if g.gameState.hold != nil {
		g.drawPiece(g.gameState.hold, x-15, y)
	}
}

func (g *Game) drawScore(width, height int) {
	x := width/2 + 10
	y := height/2 - 10

	boxStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	headStyle := tcell.StyleDefault.Background(tcell.ColorGray)
	for col := x + 1; col < x+15; col++ {
		g.screen.SetContent(col, y+8, ' ', nil, headStyle)
	}
	col := x + 6
	for _, r := range "Score" {
		g.screen.SetContent(col, y+8, r, nil, headStyle)
		col++
	}
	for row := y + 9; row < y+12; row++ {
		for col := x + 1; col < x+15; col++ {
			g.screen.SetContent(col, row, ' ', nil, boxStyle)
		}
	}

	scoreString := strconv.Itoa(g.gameState.score)
	col = x + 14 - len(scoreString)
	for _, r := range scoreString {
		g.screen.SetContent(col, y+10, r, nil, boxStyle)
		col++
	}
}

func (g *Game) drawMessage(width, height int) {
	x := width/2 - 10
	y := height/2 - 10
	style := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)

	if g.gameState.gameover || !g.gameState.started {
		msg := "Press Enter to start a new game"
		col := x - 5
		for _, r := range msg {
			g.screen.SetContent(col, y-1, r, nil, style)
			col++
		}
	} else if g.gameState.pause {
		msg := "Game Paused"
		col := x + 4
		for _, r := range msg {
			g.screen.SetContent(col, y-1, r, nil, style)
			col++
		}
	} else {
		for col := x - 5; col < x+26; col++ {
			g.screen.SetContent(col, y-1, ' ', nil, style)
		}
	}
}

func (g *Game) drawHelp(width, height int) {
	x := width/2 - 10
	y := height/2 - 10

	boxStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	headStyle := tcell.StyleDefault.Background(tcell.ColorGray)
	for col := x - 15; col < x-1; col++ {
		g.screen.SetContent(col, y+9, ' ', nil, headStyle)
	}
	col := x - 10
	for _, r := range "Help" {
		g.screen.SetContent(col, y+9, r, nil, headStyle)
		col++
	}
	for row := y + 10; row < y+17; row++ {
		for col := x - 15; col < x-1; col++ {
			g.screen.SetContent(col, row, ' ', nil, boxStyle)
		}
	}

	msgs := []string{
		"h:      left",
		"l:     right",
		"j:      down",
		"k:    rotate",
		"f:      hold",
		"space:  drop",
		"p:     pause",
	}
	for i, msg := range msgs {
		for j, r := range msg {
			g.screen.SetContent(x-14+j, y+10+i, r, nil, boxStyle)
		}
	}
}

func (g *Game) drawPiece(piece *Piece, x, y int) {
	style := piece.tile.GetStyle()
	if piece.tile == LightBlue || piece.tile == Yellow {
		x += 1
	}
	for _, loc := range piece.loc {
		g.screen.SetContent(x+(loc[1]*2)-3, y+loc[0]+5, ' ', nil, style)
		g.screen.SetContent(x+(loc[1]*2)-4, y+loc[0]+5, ' ', nil, style)
	}
}
