package model

import "github.com/gorilla/websocket"

type WebSocketConnection struct {
	*websocket.Conn
}

type MessageResponse struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type MessageRequest struct {
	Action  string              `json:"action"`
	User    string              `json:"user"`
	Message string              `json:"message"`
	Conn    WebSocketConnection `json:"-"`
}
