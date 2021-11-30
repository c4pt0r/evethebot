package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

const (
	FormatPlainText TextFormatType = iota
	FormatMarkdown
)

type Poller interface{}

type Bot interface {
	Send(chatID int64, replyMsgID int, msg string, tp TextFormatType) error
	// TODO: add poller api
}

type TgBot struct {
	bot *tgbotapi.BotAPI
}

func (b *TgBot) Send(chatID int64, replyMsgID int, m string, tp TextFormatType) error {
	msg := tgbotapi.NewMessage(chatID, m)
	switch tp {
	case FormatMarkdown:
		msg.ParseMode = "markdown"
	}
	if replyMsgID > 0 {
		msg.ReplyToMessageID = replyMsgID
	}
	_, err := b.bot.Send(msg)
	return err
}
