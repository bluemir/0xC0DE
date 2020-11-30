package server

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func (server *Server) websocket(c *gin.Context) {
	// c.Header

	websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()

		encoder := json.NewEncoder(conn)
		decoder := json.NewDecoder(conn)

		for {
			msg := map[string]interface{}{}
			if err := decoder.Decode(&msg); err != nil {
				encoder.Encode(gin.H{"msg": err.Error(), "error": true})
				return
			}
			logrus.Tracef("websocket msg: %#v", msg)
		}
	}).ServeHTTP(c.Writer, c.Request)
}
