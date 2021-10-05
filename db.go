package main

import (
	"flag"
	"log"
	"sync"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	dbPath = flag.String("db", ".eve.db", "db path")
)

type SessionModel struct {
	gorm.Model

	ChatID       int64  `gorm:"unique_index;not null"`
	Token        string `gorm:"index:idx_token"`
	From         string `gorm:"index:idx_from"`
	CreateAt     time.Time
	LastUpdateAt time.Time
}

type MessageModel struct {
	gorm.Model

	ChatID int64  `gorm:"unique_index;not null"`
	Token  string `gorm:"index:idx_token"`
	From   string `gorm:"index:idx_from"`

	MessageID string
	Text      string
	Type      string
	SendAt    time.Time
}

var (
	_once_db   sync.Once
	_global_db *gorm.DB
)

func init() {
	// FIXME
	DB()
}

func DB() *gorm.DB {
	_once_db.Do(func() {
		var err error
		_global_db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}
		_global_db.AutoMigrate(&SessionModel{})
		_global_db.AutoMigrate(&MessageModel{})
	})
	return _global_db
}

func PutOrUpdate(m *SessionModel) error {
	DB().Clauses(clause.OnConflict{DoNothing: true}).Create(m)
	return DB().Error
}
