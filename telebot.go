package main

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func newTeleBot(token string, hub2 *Hub2) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	fatalErr(err, "tgbotapi")
	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates, err := bot.GetUpdatesChan(u)
	fatalErr(err, "tgbotapi")

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			onTgCommand(&update, hub2)
		}
	}()
	return bot
}

// {
//   "ok": true,
//   "result": [
//     {
//       "update_id": 917018863,
//       "message": {
//         "message_id": 93230,
//         "from": {
//           "id": 142664361,
//           "is_bot": false,
//           "first_name": "Jiadeng",
//           "last_name": "Ruan",
//           "username": "weaming",
//           "language_code": "en"
//         },
//         "chat": {
//           "id": -339855320,
//           "title": "Instagram Flow",
//           "type": "group",
//           "all_members_are_administrators": false
//         },
//         "date": 1588859320,
//         "text": "/sub@hub2_bot",
//         "entities": [
//           {
//             "offset": 0,
//             "length": 13,
//             "type": "bot_command"
//           }
//         ]
//       }
//     }
//   ]
// }
func onTgCommand(update *tgbotapi.Update, hub2 *Hub2) {
	msg := update.Message
	cmd := msg.Command()
	args := msg.CommandArguments()
	// isGroup := strings.Contains(msg.Chat.Type, "group")

	topics := strings.Split(args, ",")
	for i, x := range topics {
		topics[i] = strings.TrimSpace(x)
	}

	if cmd != "" {
		chatid := str(msg.Chat.ID)
		userid := str(msg.From.ID)
		switch cmd {
		case "sub":
			hub2.registerTopics(chatid, userid, topics)
			hub2.subTopics(topics)
			text := fmt.Sprintf("topics you subscribed now: %v", strings.Join(hub2.TopicsOfUser(chatid, userid).Arr(), ", "))
			hub2.bot.Send(newReplyTo(msg, text))
		case "unsub":
			text := fmt.Sprintf("topics you subscribed now: %v", strings.Join(hub2.TopicsOfUser(chatid, userid).Arr(), ", "))
			hub2.bot.Send(newReplyTo(msg, text))
		default:
		}
	}
}

func newReplyTo(msg *tgbotapi.Message, text string) tgbotapi.Chattable {
	r := tgbotapi.NewMessage(msg.Chat.ID, text)
	r.ReplyToMessageID = msg.MessageID
	return r
}
