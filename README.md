A simple chat app

Design: 
- Use a Mutex.Map to keep track of Connection to server and their username -> Mutex to eliminate concurrent read and write on shared memory
- When a new connection to server, save client to map and notify client when successfully connected
- When a connection register/join room, save client username to their address in map and notify all members in chatroom about new members
- When a client broadcast message, it broadcast to all members in the chat room
- On the event a connection is dropped -> Un-register connection from the map and notify all members in the chatroom about left member

Project structure: https://github.com/golang-standards/project-layout


Command:
To run client: go run cmd/client/main.go
To run server: go run cmd/server/main.go


