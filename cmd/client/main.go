package main

import (
	"be-chat/internal/client/cli"
	"fmt"

	"github.com/gorilla/websocket"
)

func main() {
	connection, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	cli.InitializeUser(connection)

	if err != nil {
		fmt.Println(err)
	}
	go cli.ReceiveMessage(connection)
	cli.SendMessage(connection)

}
