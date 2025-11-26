package websocket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (for dev)
	},
}

func WebSocketHandler(c *gin.Context) {
	userID := c.Query("user_id")
	uid, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	AddClient(uint(uid), conn)
	defer RemoveClient(uint(uid))

	for {
		// Keep connection alive
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
