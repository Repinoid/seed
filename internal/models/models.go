package models

import (
	"log/slog"
	"time"
)

var (
	Logger     *slog.Logger
	DBEndPoint = "postgres://uname:parole@localhost:5432/dbase"
)

type UserStr struct {
	Username   string
	Meta       string
	Created_at time.Time
}
