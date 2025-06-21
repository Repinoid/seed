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

	"gomuncool/internal/handlers"
	"gomuncool/internal/models"
)

var Host = "0.0.0.0:8080"

//var models.Logger *slog.models.Logger

func main() {

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug, // Минимальный уровень логирования
		AddSource: true,            // Добавлять информацию об исходном коде

	})
	models.Logger = slog.New(handler)
	slog.SetDefault(models.Logger)

	if err := Run(); err != nil {
		models.Logger.Error("Server Shutdown by syscall", "ListenAndServe message ", err.Error())
	}
}

// run. Запуск сервера и хендлеры
func Run() (err error) {

	router := mux.NewRouter()
	router.HandleFunc("/cap", handlers.Cap).Methods("GET")

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
	models.Logger.Info("HTTP server started")

	<-done
	models.Logger.Info("Server is shutting down...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := httpServer.Shutdown(ctx); err != nil {
		models.Logger.Error("Server shutdown failed", "err is ", err.Error())
		os.Exit(1)

	}
	models.Logger.Info("Server stopped gracefully")

	return nil
}
