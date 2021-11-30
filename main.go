package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/c4pt0r/log"

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

func main() {
	flag.Parse()
	InitDB()
	bot, err := tgbotapi.NewBotAPI(getTgToken())
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = *debugMode
	if *debugMode {
		log.SetLevelByString("debug")
	}

	log.I("[Success] Authorized on account", bot.Self.UserName)

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
		sess, ok := sm.GetSessionByChatID(chatID)
		if !ok {
			sess = NewSession(chatID, update.Message.From.UserName, botWrapper)
			sm.PutSession(sess)
		}
		var msg []byte
		if *debugMode {
			msg, err = json.MarshalIndent(update.Message, "", "  ")
		} else {
			msg, err = json.Marshal(update.Message)
		}
		if err != nil {
			log.E(err)
			continue
		}
		sess.Handle(msg)
		sm.AddToUpdateQueue(sess)
	}
}
