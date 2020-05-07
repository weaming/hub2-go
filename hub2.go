// map topics to telegram chatID
package main

import (
	"encoding/json"
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
	h := Hub2{hubAPI: hubAPI, ws: nil, bot: nil, mapping: map[string]map[string]Set{}}

	h.Lock()
	h.ws = h.newHubWSConn(3)
	h.bot = h.newBot(bottoken)
	h.Unlock()

	go h.LoopOnWsResponse()
	return &h
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

func (h *Hub2) newBot(token string) *tgbotapi.BotAPI {
	return newTeleBot(token, h)
}

func (h *Hub2) LoopOnWsResponse() {
start:
	err := h.subTopics(h.Topics().Arr())
	fatalErr(err, "sub")

	for {
		msgType, message, err := h.ws.ReadMessage()
		if err != nil || msgType == websocket.CloseMessage {
			log.Println("ws read err:", err)
			h.ws.Close()

			h.Lock()
			h.ws = h.newHubWSConn(100)
			h.Unlock()
			goto start
		}
		if msgType == websocket.TextMessage {
			log.Printf("ws recv: %s", message)
		}

		msg := core.PubMessage{}
		err = json.Unmarshal(message, msg)
		if err != nil {
			log.Println("unmarshal err:", err)
			continue
		}
		switch msg.Type {
		case core.MTPlain:
		case core.MTMarkdown:
		case core.MTJSON:
		case core.MTHTML:
		case core.MTPhoto:
		case core.MTVideo:
		default:
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
	// log.Println(h.mapping)
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
		Action: core.ActionSub,
		Topics: topics,
	}
}
