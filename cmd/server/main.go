package main

import (
	"be-chat/internal/server/handlers"
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()
	http.HandleFunc("/ws", handlers.WsEndpoint)
	go handlers.HandleRequest()
	go handlers.Broadcast()
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
