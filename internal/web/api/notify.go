package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ixugo/goddd/pkg/web"
)

// socketUpgrade 函数用于将HTTP连接升级为WebSocket连接
func socketUpgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	socket := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024 * 2,
		WriteBufferSize: 1024,
	}
	return socket.Upgrade(w, r, nil)
}

func registerNotify(g gin.IRouter, handler ...gin.HandlerFunc) {
	group := g.Group("/notify")
	group.POST("/messages", func(c *gin.Context) {
		conn, err := socketUpgrade(c.Writer, c.Request)
		if err != nil {
			web.Fail(c, err)
			return
		}
		defer conn.Close()
	})
}
