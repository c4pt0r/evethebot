package main

import (
	"flag"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	debugMode = flag.Bool("debug", false, "debug mode")
)

const (
	TG_API_POLL_TIMEOUT = 60 // in second
)

func getTgToken() string {
	return os.Getenv("TOKEN")
}

type TgBot struct {
	bot *tgbotapi.BotAPI
}

func (b *TgBot) SendPlainText(chatID int64, m string) error {
	msg := tgbotapi.NewMessage(chatID, m)
	_, err := b.bot.Send(msg)
	return err
}

func (b *TgBot) SendMarkdown(chatID int64, m string) error {
	msg := tgbotapi.NewMessage(chatID, m)
	msg.ParseMode = "markdown"
	_, err := b.bot.Send(msg)
	return err
}

func main() {
	flag.Parse()

	bot, err := tgbotapi.NewBotAPI(getTgToken())
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = *debugMode

	log.Printf("[Success] Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = TG_API_POLL_TIMEOUT
	updates, _ := bot.GetUpdatesChan(u)
	botWrapper := &TgBot{bot}
	sm := NewSessionManager(botWrapper)

	httpServer := NewHttpServer(sm)
	go httpServer.Serve()
	// start polling
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if update.Message.Chat.Type != "private" {
			continue
		}
		chatID := update.Message.Chat.ID
		// get session by chat id, if not exists create one
		// TODO: add cache here
		sess, ok := sm.GetSessionByChatID(chatID)
		if !ok {
			sess = NewSession(chatID, update.Message.From.UserName, botWrapper)
			sm.PutSession(sess)
		}
		sess.Handle(update.Message.Text)
		sm.AddToUpdateQueue(sess)
	}
}
