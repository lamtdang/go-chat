package handlers

import (
	"be-chat/internal/server/model"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var messagePayload = make(chan model.MessageRequest)
var broadcastMessages = make(chan model.MessageResponse)
var clientsConnectionSync sync.Map

// Handler Controller
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := wsUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	wsConn := model.WebSocketConnection{Conn: ws}
	RegisterClientConnection(wsConn)
	go PushMessageToPayload(&wsConn)

}

// Establish first connection
func RegisterClientConnection(conn model.WebSocketConnection) {
	clientsConnectionSync.Store(conn, "")
	serverResponse := model.MessageResponse{
		User:    "Server",
		Message: "Connected to Chatroom",
	}

	conn.WriteJSON(serverResponse)
}

// Handle
func HandleRequest() {
	for {
		message := <-messagePayload
		switch message.Action {
		case "Register":
			RegisterClient(message)
		case "Broadcast":
			BroadCastUsersMessage(message)
		}

	}
}

// Receive message via WS and push to message channel
func PushMessageToPayload(conn *model.WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("RECOVER", r)
		}
	}()

	var request model.MessageRequest

	for {
		_, ok := clientsConnectionSync.Load(*conn)
		if ok {
			err := conn.ReadJSON(&request)
			if err != nil {
				UnregisterUser(conn)
			} else {
				request.Conn = *conn
				messagePayload <- request
			}
		}
	}
}

// Register new Client connection with username and list members that are already in the chat
func RegisterClient(message model.MessageRequest) {
	clientsConnectionSync.Store(message.Conn, message.User)
	var activeUsers []string
	clientsConnectionSync.Range(func(key, value any) bool {
		username := value.(string)
		if username != "" {
			activeUsers = append(activeUsers, username)
		}
		return true
	})
	sort.Strings(activeUsers)
	activeUsersStr := strings.Join(activeUsers, ", ")
	listUserStr := fmt.Sprintf("%s has just joined the chat. Currently, %s are in the chat", message.User, activeUsersStr)
	listUserMessage := model.MessageResponse{
		User:    "Server",
		Message: listUserStr,
	}
	broadcastMessages <- listUserMessage
}

// Remove Client from WS and notify all other users
func UnregisterUser(userConn *model.WebSocketConnection) {
	username, ok := clientsConnectionSync.Load(*userConn)
	if ok {
		_ = userConn.Close()
		clientsConnectionSync.Delete(*userConn)
		clientLeftNotification := model.MessageResponse{
			User:    "Server",
			Message: username.(string) + " has left the chat",
		}
		broadcastMessages <- clientLeftNotification
	}
}

// Build Response to broadcast message to all users
func BroadCastUsersMessage(message model.MessageRequest) {
	user, ok := clientsConnectionSync.Load(message.Conn)
	if ok && user != "" {
		response := model.MessageResponse{
			User:    user.(string),
			Message: message.Message,
		}
		broadcastMessages <- response
	}

}

// Broadcast response to all users
func Broadcast() {
	for {
		response := <-broadcastMessages
		clientsConnectionSync.Range(func(client, value any) bool {
			client_conversion := client.(model.WebSocketConnection)
			err := client_conversion.WriteJSON(response)
			if err != nil {
				log.Println("Error in Broadcasting to this client")

				UnregisterUser(&client_conversion)
			}
			return true
		})
	}
}
