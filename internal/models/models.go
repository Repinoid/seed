package models

import (
	"gomuncool/internal/dbase"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBstruct struct {
	//	DB     *pgx.Conn
	DB *pgxpool.Pool
}

var (
	Logger     *slog.Logger
	DBEndPoint = "postgres://uname:parole@localhost:5432/dbase"
	DataBase   *dbase.DBstruct
)
