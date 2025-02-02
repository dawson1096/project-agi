package main

import "log"

func main() {
	game := Game{}
	if err := game.Init(); err != nil {
		log.Fatalf("%v", err)
	}
	if err := game.Run(); err != nil {
		log.Fatalf("%v", err)
	}
}
