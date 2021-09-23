package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Bot interface {
	SendPlainText(chatID int64, msg string) error
	SendMarkdown(chatID int64, md string) error
}

type Message struct {
	ChatID   int64
	From     string
	To       string
	Content  string
	CreateAt time.Time
}

type Session struct {
	stopFlag   int32
	chatID     int64
	token      string
	from       string
	bot        Bot
	createAt   time.Time
	lastUpdate time.Time

	// TODO
	recentConversation []Message
}

func NewSession(chatID int64, from string, bot Bot) *Session {
	u := uuid.New()
	token := u.String()
	return &Session{
		stopFlag: 0,
		bot:      bot,
		chatID:   chatID,
		from:     from,
		token:    token,
		createAt: time.Now(),
	}
}

func (c *Session) Stop()                          { atomic.StoreInt32(&(c.stopFlag), int32(1)) }
func (c *Session) IsStop() bool                   { return atomic.LoadInt32(&(c.stopFlag)) != 0 }
func (c *Session) Token() string                  { return c.token }
func (c *Session) ChatID() int64                  { return c.chatID }
func (c *Session) From() string                   { return c.from }
func (c *Session) SendPlainText(msg string) error { return c.bot.SendPlainText(c.chatID, msg) }
func (c *Session) SendMarkdown(msg string) error  { return c.bot.SendMarkdown(c.chatID, msg) }
func (c *Session) Save() error                    { return PutOrUpdate(c.Model()) }

func (c *Session) Handle(msg string) error {
	var err error
	if msg == "/weather" {
		err = c.onWeather()
	} else if msg == "/start" {
		if err != nil {
			log.Println(err)
		}
		err = c.onUsage()
	} else if msg == "/run" {
		err = c.onRun()
	} else if msg == "/token" {
		err = c.onGetToken()
	} else if msg == "/stop" {
		c.Stop()
		c.SendPlainText("well, your call.")
	} else {
		err = c.onUsage()
	}
	c.lastUpdate = time.Now()
	return err
}

func (c *Session) onWeather() error {
	return c.SendPlainText("☀️⛈️❄️")
}

func (c *Session) onGetToken() error {
	usageStr := fmt.Sprintf(`curl -X POST http://127.0.0.1:8089/post `+
		`-d '{"token":"%s","msg":"Hello World"}'`, c.Token())
	reply := fmt.Sprintf("Your Token:\n"+c.Token()+"\nPlease don't share...😈\nHave a try:\n  %s", usageStr)
	return c.SendPlainText(reply)
}
func (c *Session) onUsage() error {
	return c.SendPlainText("Usage:\n/run\n/stop\n/token")
}

func (c *Session) onRun() error {
	return nil
}

func (s *Session) Model() *SessionModel {
	return &SessionModel{
		ChatID:   s.chatID,
		Token:    s.token,
		From:     s.from,
		CreateAt: s.createAt,
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
	return PutOrUpdate(s.Model())
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

func (sm *SessionMgr) addToUpdateQueue(s *Session) {
	// TODO may block
	sm.updateQ <- s
}

func (sm *SessionMgr) updateSessionWorker() {
	// TODO use batch
	for s := range sm.updateQ {
		s.Save()
	}
}
