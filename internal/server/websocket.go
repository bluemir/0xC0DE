package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

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

func (server *Server) stream(c *gin.Context) {
	// it is oneway... server -> client
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	gone := c.Stream(func(w io.Writer) bool {
		if t, ok := <-ticker.C; ok {
			c.SSEvent("time", t.Format(time.RFC3339Nano))
			return true
		}
		return false
	})
	if gone {
		logrus.Debug("client gone")
	}
	// stream = new EventSource("/stream")
}
func (server *Server) push(c *gin.Context) {
	pusher := c.Writer.Pusher()
	if pusher == nil {
		c.JSON(http.StatusBadRequest, "Not supported")
		return
	}
	// use web url address. it request http 1.1 request to itself, and reply it.
	// https://go.dev/blog/h2push
	if err := pusher.Push("/static/js/index.js", nil); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, "hello")
}
