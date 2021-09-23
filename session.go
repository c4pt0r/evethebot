package main

import (
	"fmt"
	"sync"
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

func (c *Session) Handle(msg string) error {
	var err error
	if msg == "/weather" {
		err = c.onWeather()
	} else if msg == "/start" {
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
	return c.SendPlainText("hello")
}

func (c *Session) onRun() error {
	go func() {
		for !c.IsStop() {
			c.SendPlainText("mock")
			c.lastUpdate = time.Now()
			time.Sleep(1 * time.Second)
		}
	}()
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

func (s *Session) Persist() error {
	return PutOrUpdate(s.Model())
}

var (
	_sessionManager *SessionMgr
	_once           sync.Once
)

func SM() *SessionMgr {
	_once.Do(func() {
		_sessionManager = &SessionMgr{
			chatTosession:  make(map[int64]*Session),
			tokenToSession: make(map[string]*Session),
		}
	})
	return _sessionManager
}

type SessionMgr struct {
	mu             sync.RWMutex
	chatTosession  map[int64]*Session
	tokenToSession map[string]*Session
}

func (sm *SessionMgr) PutSession(chatID int64, s *Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.chatTosession[chatID] = s
	sm.tokenToSession[s.Token()] = s
}

func (sm *SessionMgr) GetSessionByToken(token string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if s, ok := sm.tokenToSession[token]; ok {
		return s, true
	}
	return nil, false
}

func (sm *SessionMgr) GetSessionByChatID(chatID int64) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if s, ok := sm.chatTosession[chatID]; ok {
		return s, true
	}
	return nil, false
}

func (sm *SessionMgr) RemoveSessionByChatID(chatID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.chatTosession[chatID]; ok {
		token := s.Token()
		delete(sm.tokenToSession, token)
		delete(sm.chatTosession, chatID)
	}
}

func (sm *SessionMgr) RemoveSessionbyToken(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.tokenToSession[token]; ok {
		chatID := s.ChatID()
		delete(sm.tokenToSession, token)
		delete(sm.chatTosession, chatID)
	}
}
