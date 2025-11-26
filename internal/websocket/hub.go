package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

var Clients = make(map[uint]*websocket.Conn) // map[userID]*Conn
var Mutex = &sync.Mutex{}

func AddClient(userID uint, conn *websocket.Conn) {
	Mutex.Lock()
	defer Mutex.Unlock()
	Clients[userID] = conn
}

func RemoveClient(userID uint) {
	Mutex.Lock()
	defer Mutex.Unlock()
	delete(Clients, userID)
}

func SendNotification(userID uint, message string) {
	Mutex.Lock()
	defer Mutex.Unlock()

	conn, exists := Clients[userID]
	if !exists {
		return
	}

	err := conn.WriteJSON(map[string]string{
		"type":    "notification",
		"message": message,
	})
	if err != nil {
		conn.Close()
		delete(Clients, userID)
	}
}
