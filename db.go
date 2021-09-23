package main

import (
	"flag"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	dbPath = flag.String("db", ".eve.db", "db path")
)

type SessionModel struct {
	ChatID   int64
	Token    string
	From     string
	CreateAt time.Time
}
