package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/weaming/hub2-go/client"
)

func newTeleBot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	client.FatalErr(err, "tgbotapi")
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	client.FatalErr(err, "tgbotapi")

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			onTgCommand(&update)
		}
	}()
	return bot
}

func onTgCommand(update *tgbotapi.Update) {

}
