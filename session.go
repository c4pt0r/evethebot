package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Bot interface {
	Send(chatID int64, msg string) error
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
	bot        Bot
	createAt   time.Time
	lastUpdate time.Time

	recentConversation []Message
}

func NewSession(chatID int64, bot Bot) *Session {
	u := uuid.New()
	token := u.String()
	return &Session{
		stopFlag: 0,
		bot:      bot,
		chatID:   chatID,
		token:    token,
		createAt: time.Now(),
	}
}

func (c *Session) IsStop() bool {
	if atomic.LoadInt32(&(c.stopFlag)) != 0 {
		return true
	}
	return false
}

func (c *Session) Stop()                 { atomic.StoreInt32(&(c.stopFlag), int32(1)) }
func (c *Session) Token() string         { return c.token }
func (c *Session) ChatID() int64         { return c.chatID }
func (c *Session) Send(msg string) error { return c.bot.Send(c.chatID, msg) }

func (c *Session) Handle(msg string) error {
	var err error
	if msg == "/weather" {
		err = c.bot.Send(c.chatID, "sunny")
	} else if msg == "/start" {
		go func() {
			for !c.IsStop() {
				c.Send("mock")
				c.lastUpdate = time.Now()
				time.Sleep(1 * time.Second)
			}
		}()
	} else if msg == "/token" {
		usageStr := fmt.Sprintf(`curl -X POST http://127.0.0.1:8089/post `+
			`-d '{"token":"%s","msg":"Hello World"}'`, c.Token())
		reply := fmt.Sprintf("Your Token: "+c.Token()+"\nPlease don't share...😈\nHave a try:\n  %s", usageStr)
		err = c.Send(reply)
	} else if msg == "/stop" {
		c.Stop()
	} else {
		err = c.Send("Usage: /weathe /token /revoke_token /stop /start /status")
	}
	return err
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
