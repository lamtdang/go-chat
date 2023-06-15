package cli

import (
	"be-chat/internal/server/model"
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// Create new user and connect to chat server
func InitializeUser(conn *websocket.Conn) {
	username := ""
	for username == "" {
		fmt.Printf("Enter your username: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		username = strings.Replace(line, "\n", "", -1)
		if username == "" {
			fmt.Println("Try again")
		}
	}

	initialize_message := model.MessageRequest{
		Action: "Register",
		User:   username,
	}

	err := conn.WriteJSON(initialize_message)
	if err != nil {
		fmt.Println(err)
	}
}

// Send message from user to chat server
func SendMessage(conn *websocket.Conn) {
	for {
		fmt.Printf("-> ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		message := model.MessageRequest{
			Action:  "Broadcast",
			Message: strings.Replace(line, "\n", "", -1),
		}

		err = conn.WriteJSON(message)
		if err != nil {
			fmt.Println(err)
		}

	}
}

// Polling messages from chat server to client
func ReceiveMessage(conn *websocket.Conn) {

	var response model.MessageResponse
	for {
		err := conn.ReadJSON(&response)
		if err != nil {
			log.Println("Disconnected to server")
			os.Exit(1)
		}
		fmt.Println(response.User + ": " + response.Message)
		fmt.Printf("-> ")
	}
}
