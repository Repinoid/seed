package main

import (
	"gomuncool/internal/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var Host = ":8080"

func main() {

	if err := Run(); err != nil {
		log.Printf("Server Shutdown by syscall, ListenAndServe message -  %v\n", err)
	}
}

// run. Запуск сервера и хендлеры
func Run() (err error) {

	router := mux.NewRouter()
	router.HandleFunc("/cap", handlers.Cap).Methods("GET")

	httpServer := http.Server{Addr: Host, Handler: router}
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	return nil
}
