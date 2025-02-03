package main


type GameState struct {
	gameover bool
	totalPoints int
	board [][]int
	
}

func NewGameState(n int) *GameState {

	var boardInit [][]int

	for i := 0; i < n; i ++ {
		row := []int{}
		for j := 0; j < n; j++ {
			row[j] = 0
		}
		boardInit[i] = row
	}
	
	return &GameState{
		gameover: false,
		totalPoints: 0,
		board: boardInit,
	}
}

type Snake struct {

}

/*


Class
Track state

> Game State - class
attributes
- Board
- Points
- Game over or not
Method 
- overlap with
- did snake collide with wall

> Snake
- 2d array for coordinates (first element is head, last element )
	- Linked list?

	classes
	https://dev.to/jpoly1219/structs-methods-and-receivers-in-go-5g4f

*/