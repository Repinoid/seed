package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"gomuncool/internal/dbase"
	"gomuncool/internal/handlers"
	"gomuncool/internal/models"
)

var Host = "0.0.0.0:8080"

//var models.Logger *slog.models.Logger

func main() {
	ctx := context.Background()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Минимальный уровень логирования
		AddSource: true,            // Добавлять информацию об исходном коде
	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	// DaraBase Endpoint - if exists in Environment variable, if not - default
	enva, exists := os.LookupEnv("DATABASE_DSN")
	if exists {
		models.DBEndPoint = enva
	}

	db, err := dbase.ConnectToDB(ctx, models.DBEndPoint)
	if err != nil {
		models.Logger.Error("Can't connect to DB", "", err.Error())
		return
	}
	defer db.CloseBase()

	if err = db.UsersTableCreation(ctx); err != nil {
		models.Logger.Error("UsersTableCreation", "", err.Error())
		return
	}
	models.Logger.Debug("DB connected")
	dbase.DataBase = db

	if err := Run(ctx); err != nil {
		models.Logger.Error("Server Shutdown by syscall", "ListenAndServe message ", err.Error())
	}
}

// run. Запуск сервера и хендлеры
func Run(ctx context.Context) (err error) {

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.Cap).Methods("GET")
	router.HandleFunc("/put/{userName}/{role}", handlers.PutUser).Methods("GET")
	router.HandleFunc("/get/{userName}", handlers.GetUser).Methods("GET")

	httpServer := http.Server{Addr: Host, Handler: router}

	// Channel to listen for interrupts
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			models.Logger.Error("ListenAndServe", "err is ", err.Error())
			os.Exit(1)
		}
	}()
	models.Logger.Info("HTTP server started HAZER")

	<-done
	models.Logger.Info("Server is shutting down...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := httpServer.Shutdown(ctx); err != nil {
		models.Logger.Error("Server shutdown failed", "err is ", err.Error())
		os.Exit(1)

	}
	models.Logger.Info("Server stopped gracefully")

	return nil
}
