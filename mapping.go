// map topics to telegram chatID
package main

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
	"github.com/weaming/hub/core"
)

type Hub2 struct {
	sync.RWMutex
	ws      *websocket.Conn
	bot     *tgbotapi.BotAPI
	mapping *Mapping
}

type Mapping struct {
	m map[string][]*Receiver // topic -> list of user in one telegram group
}

type Receiver struct {
	chatid string
	userid string
}

func NewHub2(bot *tgbotapi.BotAPI, hubAPI string) *Hub2 {
	log.Printf("connecting to %s", *server)
	c, _, err := websocket.DefaultDialer.Dial(hubAPI, nil)
	fatalErr(err, "dial")

	err := h.SubTopics()
	fatalErr(err, "sub")

	return &Hub2{c, bot, &Mapping{m: map[string][]*Receiver{}}}
}

func (h *Hub2) SubTopics() error {
	return h.ws.WriteJSON(NewSubMessage(h.Topics()))
}

func (h *Hub2) Topics() (rv []string) {
	h.RLock()
	defer h.RUnlock()

	for topic := range h.mapping.m {
		rv = append(rv, topic)
	}
	return
}

func (h *Hub2) Block() {
	for {
	}
}

func NewSubMessage(topics []string) *core.PubRequest {
	return &core.PubRequest{
		Action: core.ActionPub,
		Topics: topics,
	}
}
