package main

import (
	"errors"
	"flag"
	"sync"
	"time"

	"github.com/c4pt0r/log"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gorm.io/datatypes"
)

var (
	dbPath   = flag.String("db", ".eve.db", "db path, using sqlite, for debugging only")
	mysqlDSN = flag.String("mysql", "", "mysql dsn")
)

var (
	ErrUnknownModel = errors.New("unknown model")
)

const (
	MaxFetchLimit int = 100
)

type SessionModel struct {
	gorm.Model

	ChatID       int64  `gorm:"index:,unique"`
	Token        string `gorm:"index:,unique,length:50"`
	From         string `gorm:"index:idx_session_from"`
	CreateAt     time.Time
	LastUpdateAt time.Time
}

type MessageModel struct {
	gorm.Model

	ChatID int64  `gorm:"unique_index;not null"`
	Token  string `gorm:"index:idx_message_token"`
	From   string `gorm:"index:idx_message_from"`
	Text   string

	MessageID   int64
	MessageBody datatypes.JSON
	CreateAt    time.Time `gorm:"index:idx_create_at"`
}

var (
	_once_db   sync.Once
	_global_db *gorm.DB
)

func InitDB() {
	DB()
}

func DB() *gorm.DB {
	_once_db.Do(func() {
		var err error
		if len(*mysqlDSN) > 0 {
			log.I("init mysql", *mysqlDSN)
			_global_db, err = gorm.Open(mysql.Open(*mysqlDSN), &gorm.Config{})
		} else {
			log.I("init sqlite", *dbPath)
			_global_db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
		}
		if err != nil {
			log.Fatal(err)
		}
		_global_db.AutoMigrate(&SessionModel{})
		_global_db.AutoMigrate(&MessageModel{})
	})
	return _global_db
}

func PutModel(m interface{}) error {
	switch v := m.(type) {
	case *SessionModel:
		return DB().Clauses(clause.OnConflict{UpdateAll: true}).Create(v).Error
	case *MessageModel:
		return DB().Clauses(clause.OnConflict{DoNothing: true}).Create(v).Error
	}
	return ErrUnknownModel
}

func Fetch(m interface{}, limit int, conds ...interface{}) (interface{}, error) {
	switch m.(type) {
	case MessageModel:
		var ret []MessageModel
		var err error
		if limit > 0 && limit < MaxFetchLimit {
			err = DB().Limit(limit).Order("create_at asc").Find(&ret, conds...).Error
		} else {
			err = DB().Limit(MaxFetchLimit).Order("create_at asc").Find(&ret, conds...).Error
		}
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
	return nil, ErrUnknownModel
}
