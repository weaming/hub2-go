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
	hubAPI  string
	ws      *websocket.Conn
	bot     *tgbotapi.BotAPI
	mapping map[string]map[string]Set // topic -> chatid -> username
}

func NewHub2(bottoken, hubAPI string) *Hub2 {
	hub2 := Hub2{hubAPI: hubAPI, ws: nil, bot: nil, mapping: map[string]map[string]Set{}}
	hub2.ws = hub2.newHubWSConn(3)
	hub2.newBot(bottoken)
	go hub2.loopOnWsResponse()

	err := hub2.subTopics(hub2.Topics().Arr())
	fatalErr(err, "sub")

	return &hub2
}

func (h *Hub2) newHubWSConn(retry int) *websocket.Conn {
retry:
	log.Printf("connecting to %s", h.hubAPI)
	ws, _, err := websocket.DefaultDialer.Dial(h.hubAPI, nil)
	if err != nil {
		log.Println("dial:", err)
		if retry > 0 {
			retry--
			goto retry
		}
		panic(err)
	}
	log.Printf("connectted to %s", h.hubAPI)
	return ws
}

func (h *Hub2) newBot(token string) {
	h.bot = newTeleBot(token, h)
}

func (h *Hub2) loopOnWsResponse() {
	for {
		msgType, message, err := h.ws.ReadMessage()
		if err != nil || msgType == websocket.CloseMessage {
			log.Println("ws read err:", err)

			h.Lock()
			h.ws.Close()
			h.ws = h.newHubWSConn(100)
			h.Unlock()
		}
		if msgType == websocket.TextMessage {
			log.Printf("ws recv: %s", message)
		}
	}
}

func (h *Hub2) addMapping(topic, chatid, userid string) {
	h.Lock()
	defer h.Unlock()
	if xxs, ok := h.mapping[topic]; ok {
		if xs, ok2 := xxs[chatid]; ok2 {
			xs.Add(userid)
		} else {
			xxs[chatid] = NewSet(userid)
		}
	} else {
		h.mapping[topic] = map[string]Set{chatid: NewSet(userid)}
	}
	log.Println(h.mapping)
}

func (h *Hub2) registerTopics(chatid, userid string, topics []string) {
	for _, t := range topics {
		h.addMapping(t, chatid, userid)
	}
}

func (h *Hub2) subTopics(topics []string) error {
	h.Lock()
	defer h.Unlock()
	log.Println("sub:", topics)
	return h.ws.WriteJSON(NewSubMessage(topics))
}

func (h *Hub2) Topics() Set {
	h.RLock()
	defer h.RUnlock()

	s := NewSet()
	for topic := range h.mapping {
		s.Add(topic)
	}
	return s
}

func (h *Hub2) TopicsOfUser(chatid, userid string) Set {
	h.RLock()
	defer h.RUnlock()

	s := NewSet()

	for topic, xxs := range h.mapping {
		if xs, ok := xxs[chatid]; ok {
			if xs.Has(userid) {
				s.Add(topic)
				break
			}
		}
	}
	return s
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
