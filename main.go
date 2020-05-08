package main

import (
	"flag"
	"os"
)

func main() {
	hubAPI := flag.String("ws", "wss://hub.drink.cafe/ws", "websocket server api")
	bottoken := flag.String("token", "", "telegram bot token")
	cfgPath := flag.String("config", os.ExpandEnv("$HOME/data/hub2-go/config.json"), "telegram bot token")
	flag.Parse()

	hub2 := NewHub2(*bottoken, *hubAPI, *cfgPath)
	hub2.Block()
}
