package main

import (
	"flag"
)

func main() {
	hubAPI := flag.String("ws", "wss://hub.drink.cafe/ws", "websocket server api")
	bottoken := flag.String("token", "", "telegram bot token")
	flag.Parse()

	hub2 := NewHub2(*bottoken, *hubAPI)
	hub2.Block()
}
