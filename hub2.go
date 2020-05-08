// map topics to telegram chatID
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
	"github.com/weaming/hub/core"
)

type Hub2 struct {
	sync.RWMutex            // rw lock for .mapping, .ws and .bot
	wsRLock      sync.Mutex // read lock for websocket connection
	wsWLock      sync.Mutex // write lock for websocket connection
	botLock      sync.Mutex
	HubAPI       string
	ConfigPath   string
	ws           *websocket.Conn
	bot          *tgbotapi.BotAPI
	mapping      map[string]map[string]Set // topic -> chatid -> username
}

func NewHub2(bottoken, hubAPI, configPath string) *Hub2 {
	h := Hub2{
		HubAPI:     hubAPI,
		ConfigPath: configPath,
		ws:         nil,
		bot:        nil,
		mapping:    map[string]map[string]Set{},
	}

	prepareDir(configPath, true)

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		readJson(configPath, &h.mapping)
	}

	log.Println(h.mapping)

	h.Lock()
	h.ws = h.newHubWSConn(3)
	h.bot = h.newBot(bottoken)
	h.Unlock()

	go h.LoopOnWsResponse()
	return &h
}

// try connecting to Hub websocket, panic after use out of retry limit
func (h *Hub2) newHubWSConn(retry int) *websocket.Conn {
retry:
	log.Printf("connecting to %s", h.HubAPI)
	ws, _, err := websocket.DefaultDialer.Dial(h.HubAPI, nil)
	if err != nil {
		log.Println("dial:", err)
		if retry > 0 {
			retry--
			goto retry
		}
		panic(err)
	}
	log.Printf("connectted to %s", h.HubAPI)
	return ws
}

func (h *Hub2) newBot(token string) *tgbotapi.BotAPI {
	return newTeleBot(token, h)
}

func (h *Hub2) LoopOnWsResponse() {
	defer func() { log.Println("read loop exited") }()

start:
	err := h.subTopics(h.Topics().Arr())
	fatalErr(err, "sub")

	for {
		msgType, message, err := func() (int, []byte, error) {
			h.wsRLock.Lock()
			defer h.wsRLock.Unlock()

			return h.ws.ReadMessage()
		}()

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

		push := &core.PushMessageResponse{}
		err = json.Unmarshal(message, push)
		if err != nil {
			log.Println("unmarshal err:", err)
			continue
		}

		// go h.handlerWSPush(*push, message) // as write-lock botLock in handlerWSPush, new goroutine is useless
		h.handlerWSPush(*push, message)
	}
}

func (h *Hub2) handlerWSPush(push core.PushMessageResponse, message []byte) {
	// add read lock for h.mapping
	h.RLock()
	defer h.RUnlock()

	// add write lock for h.bot.Send
	// Note: comment it will not work properly
	h.botLock.Lock()
	defer h.botLock.Unlock()

	switch push.Type {
	case core.MTResponse, core.MTFeedback:
		// pass
	case core.MTMessage:
		push := &core.PushMessage{}

		err := json.Unmarshal(message, push)
		if err != nil {
			log.Println("unmarshal err:", err)
			return
		}

		msg := push.Message

		body := ""
		bodyArr := []core.RawItem{}
		preview := false
		mode := ""

		switch msg.Type {
		case core.MTPlain:
			body = msg.Data
		case core.MTJSON:
			body = msg.Data
		case core.MTMarkdown:
			body = msg.Data
			mode = "Markdown"
		case core.MTMarkdownV2:
			body = msg.Data
			mode = "MarkdownV2"
		case core.MTHTML:
			body = msg.Data
			mode = "HTML"
		case core.MTPhoto, core.MTVideo:
			bodyArr = msg.ExtendedData
			bodyArr = append(bodyArr, core.RawItem{
				Type:    msg.Type,
				Data:    msg.Data,
				Caption: msg.Caption,
				Preview: msg.Preview,
			})
		default:
			log.Printf("unknown type %v\n", msg.Type)
			return
		}

		if xxs, ok := h.mapping[push.Topic]; ok {
			// log.Println(xxs)

			for chatid, useridSet := range xxs {
				chatidInt64, err := strconv.ParseInt(chatid, 10, 64)
				fatalErr(err, "strconv")

				isGroup := chatidInt64 < 0

				if len(bodyArr) == 0 {
					// send one msg
					text := fmt.Sprintf("%s\n\n# %s", body, push.Topic)
					if isGroup {
						text += fmt.Sprintf(" by %s", strings.Join(useridSet.Arr(), ", "))
					}

					tgmsg := tgbotapi.NewMessage(chatidInt64, text)
					tgmsg.DisableWebPagePreview = !preview
					tgmsg.ParseMode = mode

					// log.Println("sending...")
					_, err = h.bot.Send(tgmsg)
					if err != nil {
						// TODO
						log.Println("botsend:", err)
					}
				} else {
					// send collection of photo, video
					files := []interface{}{}
					for _, x := range bodyArr {
						switch x.Type {
						case core.MTPhoto:
							media := tgbotapi.NewInputMediaPhoto(x.Data)
							media.Caption = x.Caption
							media.ParseMode = ""
							files = append(files, media)
						case core.MTVideo:
							media := tgbotapi.NewInputMediaVideo(x.Data)
							media.Caption = x.Caption
							media.ParseMode = ""
							files = append(files, media)
						default:
							log.Printf("unknown message type %s\n", x.Type)
						}
					}

					tgmsg := tgbotapi.NewMediaGroup(chatidInt64, files)

					// log.Println("sending...")
					res, err := h.bot.Send(tgmsg)
					if err != nil {
						// TODO
						log.Println("botsend:", err, res)
					}
				}
			}
		}
	default:
		log.Printf("unknown push type %s\n", push.Type)
	}
}

func (h *Hub2) addMapping(topic, chatid, userid string) {
	if xxs, ok := h.mapping[topic]; ok {
		if xs, ok2 := xxs[chatid]; ok2 {
			xs.Add(userid)
		} else {
			xxs[chatid] = NewSet(userid)
		}
	} else {
		h.mapping[topic] = map[string]Set{chatid: NewSet(userid)}
	}
}

func (h *Hub2) registerTopics(chatid, userid string, topics []string) {
	// add write lock for update h.mapping
	h.Lock()
	defer h.Unlock()

	for _, t := range topics {
		h.addMapping(t, chatid, userid)
	}

	// write updated mapping to disk
	writeJson(h.ConfigPath, h.mapping)
}

func (h *Hub2) subTopics(topics []string) error {
	// ensure only one write to websocket
	h.wsWLock.Lock()
	defer h.wsWLock.Unlock()

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
