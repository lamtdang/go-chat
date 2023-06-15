package handlers

import (
	"be-chat/internal/server/model"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func setUpServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(WsEndpoint))
	go HandleRequest()
	go Broadcast()
	return server
}

func setUpClient(server *httptest.Server, username string) *websocket.Conn {
	var err1 error
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client, _, err1 := websocket.DefaultDialer.Dial(wsURL+"/ws", nil)
	if err1 != nil {
		log.Fatal(err1)
	}

	client = registerClient(client, username)
	return client
}

func registerClient(conn *websocket.Conn, username string) *websocket.Conn {
	input := model.MessageRequest{
		Action: "Register",
		User:   username}
	err := conn.WriteJSON(input)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		var response model.MessageResponse
		err := conn.ReadJSON(&response)
		if err != nil {
			log.Fatal(err)
		}
	}
	return conn
}

func teardownClient(conn *websocket.Conn) {
	conn.Close()
}

func teardownServer(server *httptest.Server) {
	server.Close()
}

func TestInitializeConnection(t *testing.T) {
	// Set up Server
	server := setUpServer()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	person1, _, err1 := websocket.DefaultDialer.Dial(wsURL+"/ws", nil)
	if err1 != nil {
		log.Fatal(err1)
	}

	// Test First connection to Server
	t.Run("Initialize Person 1 Connection", func(t *testing.T) {
		var ans model.MessageResponse
		err := person1.ReadJSON(&ans)
		if err != nil {
			t.Error(err)
		}
		want := model.MessageResponse{
			User:    "Server",
			Message: "Connected to Chatroom",
		}
		if ans != want {
			t.Errorf("got %s, want %s", ans, want)
		}
	})

	// Test Register new Connection
	t.Run("Test Register Connection", func(t *testing.T) {

		input := model.MessageRequest{
			Action: "Register",
			User:   "person 1"}

		want := model.MessageResponse{
			User:    "Server",
			Message: fmt.Sprintf("%s has just joined the chat. Currently, %s are in the chat", "person 1", "person 1"),
		}

		writeErr := person1.WriteJSON(input)
		if writeErr != nil {
			t.Error(writeErr)
		}

		var ans model.MessageResponse
		readErr := person1.ReadJSON(&ans)
		if readErr != nil {
			t.Error(readErr)
		}

		if ans != want {
			t.Errorf("got %s, want %s", ans, want)
		}
	})
	t.Cleanup(func() {
		teardownServer(server)
		teardownClient(person1)
	})

}

func TestNewClientJoinRoom(t *testing.T) {
	server := setUpServer()
	// Initial Client
	person1 := setUpClient(server, "person 1")

	// New Client joining the room
	person2 := setUpClient(server, "person 2")
	var ans model.MessageResponse

	// Expect notification from person 1
	err := person1.ReadJSON(&ans)
	if err != nil {
		t.Error(err)
	}
	want := model.MessageResponse{
		User:    "Server",
		Message: fmt.Sprintf("%s has just joined the chat. Currently, %s are in the chat", "person 2", "person 1, person 2"),
	}
	t.Run("Test New Client Join Room", func(t *testing.T) {
		if ans != want {
			t.Errorf("got %s, want %s", ans, want)
		}
	})

	// Tear Down
	t.Cleanup(func() {
		teardownClient(person1)
		teardownClient(person2)
		teardownServer(server)
	})

}

func TestClientLeaveRoom(t *testing.T) {

	// Set up
	server := setUpServer()
	person1 := setUpClient(server, "person 1")

	// New Client joining the room
	person2 := setUpClient(server, "person 2")
	want := model.MessageResponse{
		User:    "Server",
		Message: fmt.Sprintf("%s has left the chat", "person 1"),
	}

	// Person1 leave room
	teardownClient(person1)

	// Person 2 get notification about person 1 leaving
	var ans model.MessageResponse
	err := person2.ReadJSON(&ans)
	if err != nil {
		t.Error(err)
	}
	t.Run("Test Client Leave Room", func(t *testing.T) {
		if ans != want {
			t.Errorf("got %s, want %s", ans, want)
		}
	})

	//tear down
	t.Cleanup(func() {
		teardownClient(person2)
		teardownServer(server)
	})

}

func TestBroadcastMessage(t *testing.T) {
	//Set Up
	server := setUpServer()
	sender := setUpClient(server, "sender")
	receiver := setUpClient(server, "receiver")

	t.Run("Test broadcast messages to all connected users", func(t *testing.T) {

		// skip a message
		var response_1 model.MessageResponse
		readErr1 := sender.ReadJSON(&response_1)
		if readErr1 != nil {
			t.Error(readErr1)
		}

		input := model.MessageRequest{
			Action:  "Broadcast",
			User:    "sender",
			Message: "Successful Broadcast"}

		want := model.MessageResponse{
			User:    "sender",
			Message: "Successful Broadcast",
		}

		// Send 10000 messages to 2 subsciber
		for i := 0; i < 10000; i++ {
			writeErr := sender.WriteJSON(input)
			if writeErr != nil {
				t.Error(writeErr)
			}

			var response_1 model.MessageResponse
			readErr1 := sender.ReadJSON(&response_1)
			if readErr1 != nil {
				t.Error(readErr1)
			}

			var response_2 model.MessageResponse
			readErr2 := receiver.ReadJSON(&response_2)
			if readErr2 != nil {
				t.Error(readErr2)
			}
			if response_1 != want {
				t.Errorf("got %s, want %s", readErr1, want)
			}

			if response_2 != want {
				t.Errorf("got %s, want %s", readErr2, want)
			}
		}
	})

	//Tear Down
	t.Cleanup(func() {
		teardownClient(sender)
		teardownClient(receiver)
		teardownServer(server)
	})
}
