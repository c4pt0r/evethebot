package main

import (
	"flag"
	"log"
	"sync"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	dbPath = flag.String("db", ".eve.db", "db path")
)

type SessionModel struct {
	gorm.Model

	ChatID   int64 ``
	Token    string
	From     string
	CreateAt time.Time
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
	})
	return _global_db
}

func PutOrUpdate(m *SessionModel) error {
	return nil
}
