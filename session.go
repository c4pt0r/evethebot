package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		if msg == "/start" {
			err = c.onUsage()
		} else if msg == "/token" {
			err = c.onGetToken()
		} else if msg == "/help" {
			err = c.onUsage()
		} else {
			c.SendPlainText("OK")
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
func (c *Session) GetMessages(limit int) []MessageModel {
	msgs, err := Fetch(MessageModel{}, limit, "token = ?", c.token)
	if err != nil {
		log.E(err)
		return nil
	}
	return msgs.([]MessageModel)
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
	return c.SendPlainText("Usage: /token")
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

type SessionMgr struct {
	bot     Bot
	updateQ chan *Session
}

func NewSessionManager(bot Bot) *SessionMgr {
	mgr := &SessionMgr{
		bot:     bot,
		updateQ: make(chan *Session, 100),
	}
	go mgr.updateSessionWorker()
	return mgr
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
		log.D("save session")
		s.Save()
	}
}
