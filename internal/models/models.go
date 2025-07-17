package models

import (
	"log/slog"
)

var (
	Logger     *slog.Logger
	DBEndPoint = "postgres://uname:parole@localhost:5432/dbase"
)
