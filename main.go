package main

import (
	"big2/server"
	"big2/web_socket"
	"log"
	"net/http"
)

// Add these error types
type ErrNotPlayerTurn struct {
	Msg string
}

func (e ErrNotPlayerTurn) Error() string {
	return e.Msg
}

type ErrInvalidPlay struct {
	Msg string
}

func (e ErrInvalidPlay) Error() string {
	return e.Msg
}

func main() {
	port := "8080"
	server := server.NewServer()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		web_socket.HandleWebSocket(server, w, r)
	})

	log.Println("Starting server on :", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
