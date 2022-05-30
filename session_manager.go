package main

import "github.com/c4pt0r/log"

type SessionMgr struct {
	bot     Bot
	hive    *Hive
	updateQ chan *Session
}

func NewSessionManager(bot Bot) *SessionMgr {
	mgr := &SessionMgr{
		bot:     bot,
		hive:    NewHive(),
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
		sm:         sm,
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
