package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/c4pt0r/log"
	"github.com/google/uuid"
)

type Bot interface {
	SendPlainText(chatID int64, msg string) error
	SendMarkdown(chatID int64, md string) error
}

type Session struct {
	stopFlag   int32
	chatID     int64
	token      string
	from       string
	bot        Bot
	createAt   time.Time
	lastUpdate time.Time
}

func NewSession(chatID int64, from string, bot Bot) *Session {
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
	}
}

func (c *Session) Stop()                          { atomic.StoreInt32(&(c.stopFlag), int32(1)) }
func (c *Session) IsStop() bool                   { return atomic.LoadInt32(&(c.stopFlag)) != 0 }
func (c *Session) Token() string                  { return c.token }
func (c *Session) ChatID() int64                  { return c.chatID }
func (c *Session) From() string                   { return c.from }
func (c *Session) SendPlainText(msg string) error { return c.bot.SendPlainText(c.chatID, msg) }
func (c *Session) SendMarkdown(msg string) error  { return c.bot.SendMarkdown(c.chatID, msg) }
func (c *Session) Save() error                    { return PutModel(c.Model()) }

func (c *Session) Handle(msgJson []byte) error {
	m := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewReader(msgJson))
	decoder.UseNumber()
	decoder.Decode(&m)

	var err error
	data, ok := m["text"]
	if ok {
		msg := data.(string)
		if msg == "/start" {
			err = c.onUsage()
		} else if msg == "/run" {
			err = c.onRun()
		} else if msg == "/token" {
			err = c.onGetToken()
		} else if msg == "/stop" {
			c.Stop()
			c.SendPlainText("not implemented")
		} else if msg == "/help" {
			err = c.onUsage()
		} else {
			err = c.onUsage()
		}
	}
	c.lastUpdate = time.Now()
	msgId, ok := m["message_id"]
	if !ok {
		log.E("messageID not found, shouldn't be here")
	}
	i, err := msgId.(json.Number).Int64()
	if err != nil {
		return err
	}

	text, _ := m["text"]
	err = c.putMessage(i, text.(string), msgJson)
	return err
}

func (c *Session) onWeather() error {
	return c.SendPlainText("☀️⛈️❄️")
}

func (c *Session) onGetToken() error {
	usageStr := fmt.Sprintf(`curl -X POST `+*advisoryAddr+`/post `+
		`-d '{"token":"%s","msg":"*Hello* World"}'`, c.Token())
	reply := fmt.Sprintf("Your Token:\n"+c.Token()+"\nPlease don't share...😈\nHave a try:\n  %s", usageStr)
	return c.SendPlainText(reply)
}

func (c *Session) onUsage() error {
	return c.SendPlainText("Usage:\n/run\n/stop\n/token")
}

func (c *Session) onRun() error {
	c.SendPlainText("not implemented")
	return nil
}

func (c *Session) putMessage(messageID int64, text string, messageBody []byte) error {
	mm := &MessageModel{
		ChatID:      c.chatID,
		Token:       c.token,
		From:        c.from,
		MessageID:   messageID,
		Text:        text,
		MessageBody: string(messageBody),
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

type SessionMgr struct {
	bot     Bot
	updateQ chan *Session
}

func NewSessionManager(bot Bot) *SessionMgr {
	return &SessionMgr{
		bot:     bot,
		updateQ: make(chan *Session, 100),
	}
}

func (sm *SessionMgr) sessionModelToSessionObj(model *SessionModel) *Session {
	return &Session{
		chatID:     model.ChatID,
		from:       model.From,
		token:      model.Token,
		createAt:   model.CreateAt,
		lastUpdate: model.LastUpdateAt,
		bot:        sm.bot,
	}
}

func (sm *SessionMgr) PutSession(s *Session) error {
	return PutModel(s.Model())
}

func (sm *SessionMgr) GetSessionByToken(token string) (*Session, bool) {
	var model SessionModel
	DB().First(&model, "token = ?", token)
	if DB().Error != nil {
		return nil, false
	}
	return sm.sessionModelToSessionObj(&model), true
}

func (sm *SessionMgr) GetSessionByChatID(chatID int64) (*Session, bool) {
	var model SessionModel
	db := DB().First(&model, "chat_id = ?", chatID)
	if db.Error != nil {
		return nil, false
	}
	return sm.sessionModelToSessionObj(&model), true
}

func (sm *SessionMgr) AddToUpdateQueue(s *Session) {
	// TODO may block
	sm.updateQ <- s
}

func (sm *SessionMgr) updateSessionWorker() {
	// TODO use batch
	for s := range sm.updateQ {
		s.Save()
	}
}
