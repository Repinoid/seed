package dbase

import (
	"context"
	"fmt"
	"gomuncool/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBstruct struct {
	//	DB     *pgx.Conn
	DB *pgxpool.Pool
}

var DataBase *DBstruct

// ConnectToDB получить эндпоинт Базы Данных
func ConnectToDB(ctx context.Context, DBEndPoint string) (dataBase *DBstruct, err error) {
	//	baza, err := pgx.Connect(ctx, DBEndPoint)
	baza, err := pgxpool.New(ctx, DBEndPoint)
	if err != nil {
		return nil, fmt.Errorf("pgx can't connect to DB %s err %w", DBEndPoint, err)
	}
	// pgx.Connect возвращает err nil даже если базы не существует. так что пингуем
	err = baza.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping. can't connect to DB %s err %w", DBEndPoint, err)
	}
	dataBase = &DBstruct{DB: baza} // Initialize
	models.Logger.Debug("DB connected")

	return
}

// UsersTableCreation создание таблицы юзеров
func (dataBase *DBstruct) UsersTableCreation(ctx context.Context) error {

	db := dataBase.DB
	_, err := db.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto;") // расширение для хэширования паролей
	if err != nil {
		models.Logger.Error("error CREATE EXTENSION pgcrypto")
		return fmt.Errorf("error CREATE EXTENSION pgcrypto; %w", err)
	}

	creatorOrder :=
		"CREATE TABLE IF NOT EXISTS USERA" +
			"(userId INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, " +
			"username VARCHAR(64) UNIQUE, " +
			"metadata TEXT, " +
			"user_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"

	_, err = db.Exec(ctx, creatorOrder)
	if err != nil {
		models.Logger.Error("error CREATE USERA table")
		return fmt.Errorf("create USERS table. %w", err)
	}
	models.Logger.Debug("USERA table is created")
	return nil
}

func (dataBase *DBstruct) CloseBase() {
	dataBase.DB.Close()
}

// GetUser возвращает роль юзера
func (dataBase *DBstruct) GetUser(ctx context.Context, uname string) (meta string, err error) {
	order := "SELECT metadata FROM USERA WHERE username = $1;"
	row := dataBase.DB.QueryRow(ctx, order, uname)
	err = row.Scan(&meta)
	if err != nil {
		return "", fmt.Errorf("unknown user %+v", uname)
	}
	models.Logger.Debug("GET user ok")
	return

}

func (dataBase *DBstruct) PutUser(ctx context.Context, uname, meta string) (err error) {

	order := "INSERT INTO USERA AS args(username, metadata) VALUES ($1,$2)" +
		"ON CONFLICT (username) DO UPDATE SET username=args.username, metadata=args.metadata;"
	// args.value - старое значение. EXCLUDED.value - новое, переданное для вставки или обновления
	_, err = dataBase.DB.Exec(ctx, order, uname, meta)
	if err != nil {
		return fmt.Errorf("error insert/update %+v error is %w", uname, err)
	}
	models.Logger.Debug(" user ok")
	return

}
