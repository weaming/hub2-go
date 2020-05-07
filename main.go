package main

import (
	"flag"
)

func main() {
	hubAPI := flag.String("ws", "wss://hub.drink.cafe/ws", "websocket server api")
	bottoken := flag.String("token", "", "telegram bot token")
	flag.Parse()

	bot := newTeleBot(*bottoken)
	hub2 := NewHub2(bot, *hubAPI)
	hub2.Block()
}
