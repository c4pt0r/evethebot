package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/c4pt0r/log"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type TextFormatType int

type Session struct {
	stopFlag   int32
	chatID     int64
	token      string
	from       string
	bot        Bot
	createAt   time.Time
	lastUpdate time.Time

	sm *SessionMgr
}

func NewSession(sm *SessionMgr, chatID int64, from string, bot Bot) *Session {
	u := uuid.New()
	token := u.String()
	return &Session{
		stopFlag:   0,
		chatID:     chatID,
		token:      token,
		from:       from,
		bot:        bot,
		createAt:   time.Now(),
		lastUpdate: time.Now(),
		sm:         sm,
	}
}

func (c *Session) Stop()         { atomic.StoreInt32(&(c.stopFlag), int32(1)) }
func (c *Session) IsStop() bool  { return atomic.LoadInt32(&(c.stopFlag)) != 0 }
func (c *Session) Token() string { return c.token }
func (c *Session) ChatID() int64 { return c.chatID }
func (c *Session) From() string  { return c.from }
func (c *Session) SendPlainText(msg string) error {
	return c.bot.Send(c.chatID, 0, msg, FormatPlainText)
}
func (c *Session) SendMarkdown(msg string) error { return c.bot.Send(c.chatID, 0, msg, FormatMarkdown) }
func (c *Session) Save() error                   { return PutModel(c.Model()) }

func (c *Session) Handle(msgJson []byte) error {
	m := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewReader(msgJson))
	decoder.UseNumber()
	decoder.Decode(&m)

	var err error
	data, ok := m["text"]

	if ok {
		msg := data.(string)
		if msg[0] == '/' {
			cmd := strings.SplitN(msg, " ", 2)[0]
			if cmd == "/start" {
				err = c.onUsage()
			} else if cmd == "/token" {
				err = c.onGetToken()
			} else if cmd == "/help" {
				err = c.onUsage()
			} else if cmd == "/bees" {
				err = c.onGetBees()
			} else if cmd == "/to" {
				err = c.onToBee(m)
			} else {
				c.SendPlainText("Unknown command: " + msg + "\nTry /help")
			}
		}
	}
	c.lastUpdate = time.Now()
	msgId, ok := m["message_id"]
	if !ok {
		log.E("messageID not found, shouldn't be here")
	}
	messageID, err := msgId.(json.Number).Int64()
	if err != nil {
		return err
	}
	return c.putMessage(messageID, data.(string), msgJson)
}

// TODO: add filters, like limit/column
func (c *Session) GetMessages(limit int, offset int) []MessageModel {
	msgs, err := Fetch(MessageModel{}, limit, "token = ? and message_id >= ?", c.token, offset)
	if err != nil {
		log.E(err)
		return nil
	}
	return msgs.([]MessageModel)
}

func (c *Session) putMessage(messageID int64, text string, messageBody []byte) error {
	mm := &MessageModel{
		ChatID:      c.chatID,
		Token:       c.token,
		From:        c.from,
		MessageID:   messageID,
		Text:        text,
		MessageBody: datatypes.JSON(messageBody),
		CreateAt:    time.Now(),
	}
	return PutModel(mm)
}

func (s *Session) Model() *SessionModel {
	return &SessionModel{
		ChatID:       s.chatID,
		Token:        s.token,
		From:         s.from,
		CreateAt:     s.createAt,
		LastUpdateAt: s.lastUpdate,
	}
}

func (c *Session) onGetToken() error {
	usageStr := fmt.Sprintf(`curl -X POST `+*advisoryAddr+`/post `+
		`-d '{"token":"%s","msg":"*Hello* World"}'`, c.Token())
	reply := fmt.Sprintf("Your Token:\n`"+c.Token()+"`\nPlease don't share...😈\nHave a try:\n`%s`", usageStr)
	return c.SendMarkdown(reply)
}

func (c *Session) onUsage() error {
	return c.SendPlainText("Usage: /token /to /bees /help")
}

func (c *Session) onGetBees() error {
	bees := c.sm.hive.AllBees()
	if len(bees) == 0 {
		return c.SendPlainText("No bees found")
	}
	var buf bytes.Buffer
	for _, b := range bees {
		buf.WriteString(fmt.Sprintf("%s:%s\n", b.BeeName, b.InstanceID))
		buf.WriteString(fmt.Sprintf("last heartbeat: %s\n", b.GetLastHeartbeat()))
	}
	return c.SendPlainText(buf.String())
}

func (c *Session) onToBee(m map[string]interface{}) error {
	data, ok := m["text"]
	if !ok {
		return c.SendPlainText("Usage: /to <bee-instance-id> <message>...")
	}
	msg := data.(string)
	args := strings.SplitN(msg, " ", 3)
	if len(args) != 3 {
		return c.SendPlainText("Usage: /to <bee-instance-id> <message>...")
	}
	beesInstanceID := args[1]
	if err := c.sm.hive.SendToBee(beesInstanceID, args[2]); err != nil {
		return c.SendPlainText("Error: " + err.Error())
	}
	return c.SendPlainText("Message sent to " + beesInstanceID)
}
