package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getTgToken() string {
	return os.Getenv("TOKEN")
}

type TgBot struct {
	bot *tgbotapi.BotAPI
}

func (b *TgBot) Send(chatID int64, m string) error {
	msg := tgbotapi.NewMessage(chatID, m)
	_, err := b.bot.Send(msg)
	return err
}

func main() {
	bot, err := tgbotapi.NewBotAPI(getTgToken())
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("[Success] Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	botWrapper := &TgBot{bot}

	go serveHttp()

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s|%s] %s", update.Message.Chat.Type, update.Message.From.UserName, update.Message.Text)

		// ignore group message
		if update.Message.Chat.Type != "private" {
			continue
		}

		chatID := update.Message.Chat.ID

		// get session by chat id, if not exists create one
		sess, ok := SM().GetSessionByChatID(chatID)
		if !ok {
			sess = NewSession(chatID, botWrapper)
			SM().PutSession(sess.chatID, sess)
		}

		sess.Handle(update.Message.Text)

		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, conv.Token()+"\n Don't share to others, 🤐")
	}
}
