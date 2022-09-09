package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
)

func TestClient_BadHandshake(t *testing.T) {
	srv := newWsServer(false)
	defer srv.close()

	dialer := WebSocketDialer{Url: srv.url("/")}
	client := NewClient(dialer)
	err := client.Connect()

	assert.Error(t, errors.New("websocket: bad handshake"), err)
}

func TestClient_ServerDisconnect(t *testing.T) {
	srv := newWsServer(false)
	srv.closeOnNextMessage.Store(true)
	defer srv.close()

	dialer := WebSocketDialer{Url: srv.url("/ws")}
	client := NewClient(dialer)
	err := client.Connect()
	assert.Nil(t, err)

	req, err := NewRequest("ping", nil)
	assert.Nil(t, err)

	resp, err := client.Send(req)
	assert.Nil(t, resp)
	assert.Error(t, ErrClosed, err)
}

func TestClient_RequestIdMatching(t *testing.T) {
	srv := newWsServer(false)
	defer srv.close()

	dialer := WebSocketDialer{Url: srv.url("/ws")}
	client := NewClient(dialer)
	err := client.Connect()
	assert.Nil(t, err)

	for i := 0; i < 1000; i++ {
		ping := newRequest("ping", nil, RequestNumericId(i))
		assert.Nil(t, err)

		pong := newResponse("pong", ResponseNumericId(i))
		assert.Nil(t, err)

		pongBytes, err := json.Marshal(pong)
		assert.Nil(t, err)
		srv.testMessages <- testMessage{msgType: websocket.TextMessage, data: pongBytes}

		resp, err := client.Send(ping)
		assert.Nil(t, err)
		assert.Equal(t, pong, resp)
	}
}

func TestClient_RequestHandling(t *testing.T) {
	srv := newWsServer(true)
	defer srv.close()

	dialer := WebSocketDialer{Url: srv.url("/ws")}
	client := NewClient(dialer)

	requests := make(chan Request, 16)
	client.SetRequestHandler(func(req Request) {
		requests <- req
	})

	err := client.Connect()
	assert.Nil(t, err)

	var expected []Request
	var received []Request

	for i := 0; i < 1000; i++ {
		req := newRequest("ping", nil, RequestNumericId(i))
		expected = append(expected, req)

		bytes, err := json.Marshal(req)
		assert.Nil(t, err)

		srv.testMessages <- testMessage{msgType: websocket.TextMessage, data: bytes}
		received = append(received, <-requests)
	}

	assert.Equal(t, expected, received)
}
