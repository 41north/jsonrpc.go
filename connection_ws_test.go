package jsonrpc_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

type testMessage struct {
	msgType int
	data    []byte
}

func newWsServer(push bool) *wsServer {
	srv := wsServer{}
	srv.push = push
	srv.start()
	return &srv
}

type wsServer struct {
	srv                *httptest.Server
	testMessages       chan testMessage
	push               bool
	closeOnNextMessage atomic.Bool
}

func (t *wsServer) start() {
	t.srv = httptest.NewServer(t)
	t.testMessages = make(chan testMessage, 16)
}

func (t *wsServer) close() {
	t.srv.Close()
}

func (t *wsServer) url(path string) string {
	return strings.Replace(t.srv.URL, "http", "ws", 1) + path
}

func (t *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method != http.MethodGet {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	if path != "/ws" {
		// return a 200 response indicating no upgrade is available
		w.Write([]byte{})
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		if t.push {
			// notification testing
			for msg := range t.testMessages {
				err := c.WriteMessage(msg.msgType, msg.data)
				if err != nil {
					log.WithError(err).Error("failed to write message")
					return
				}
			}
		} else {

			// normal request -> response

			_, _, err := c.ReadMessage()
			if err != nil {
				log.WithError(err).Error("failed to read message")
				return
			}

			if t.closeOnNextMessage.Load() {
				return
			}

			msg := <-t.testMessages
			err = c.WriteMessage(msg.msgType, msg.data)
			if err != nil {
				log.WithError(err).Error("failed to write message")
				return
			}
		}
	}
}
