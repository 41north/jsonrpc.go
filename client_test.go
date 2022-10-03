package jsonrpc_test

import (
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/41north/jsonrpc.go"

	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
)

func TestClient_BadHandshake(t *testing.T) {
	srv := newWsServer(false)
	defer srv.close()

	dialer := jsonrpc.WebSocketDialer{Url: srv.url("/")}
	client := jsonrpc.NewClient(dialer)
	err := client.Connect()

	assert.Error(t, errors.New("websocket: bad handshake"), err)
}

func TestClient_ServerDisconnect(t *testing.T) {
	srv := newWsServer(false)
	srv.closeOnNextMessage.Store(true)
	defer srv.close()

	dialer := jsonrpc.WebSocketDialer{Url: srv.url("/ws")}
	client := jsonrpc.NewClient(dialer)

	// capture close errors
	closeError := atomic.Value{}
	client.SetCloseHandler(func(err error) {
		closeError.Store(err)
	})

	err := client.Connect()
	assert.Nil(t, err)

	req, err := jsonrpc.NewRequest("ping", nil)
	assert.Nil(t, err)

	var resp jsonrpc.Response
	err = client.Send(*req, &resp)
	assert.Error(t, jsonrpc.ErrClosed, err)
	assert.Equal(t, jsonrpc.ErrClosed, closeError.Load())
}

func TestClient_RequestIdMatching(t *testing.T) {
	srv := newWsServer(false)
	defer srv.close()

	dialer := jsonrpc.WebSocketDialer{Url: srv.url("/ws")}
	client := jsonrpc.NewClient(dialer)
	err := client.Connect()
	assert.Nil(t, err)

	for i := 0; i < 1000; i++ {
		ping := newRequest("ping", nil, jsonrpc.RequestNumericId(i))
		assert.Nil(t, err)

		pong := newResponse("pong", jsonrpc.ResponseNumericId(i))
		assert.Nil(t, err)

		pongBytes, err := json.Marshal(pong)
		assert.Nil(t, err)
		srv.testMessages <- testMessage{msgType: websocket.TextMessage, data: pongBytes}

		var resp jsonrpc.Response
		err = client.Send(*ping, &resp)
		assert.Nil(t, err)
		assert.Equal(t, *pong, resp)
	}
}

func TestClient_RequestHandling(t *testing.T) {
	srv := newWsServer(true)
	defer srv.close()

	dialer := jsonrpc.WebSocketDialer{Url: srv.url("/ws")}
	client := jsonrpc.NewClient(dialer)

	requests := make(chan jsonrpc.Request, 16)
	client.SetRequestHandler(func(req jsonrpc.Request) {
		requests <- req
	})

	err := client.Connect()
	assert.Nil(t, err)

	var expected []jsonrpc.Request
	var received []jsonrpc.Request

	for i := 0; i < 1000; i++ {
		req := newRequest("ping", nil, jsonrpc.RequestNumericId(i))
		expected = append(expected, *req)

		bytes, err := json.Marshal(req)
		assert.Nil(t, err)

		srv.testMessages <- testMessage{msgType: websocket.TextMessage, data: bytes}
		received = append(received, <-requests)
	}

	assert.Equal(t, expected, received)
}

// newRequest is an internal test utility for creating request objects without having to handle
// the possible error, panicking instead.
func newRequest(method string, params any, options ...jsonrpc.RequestOption) *jsonrpc.Request {
	r, err := jsonrpc.NewRequest(method, params, options...)
	if err != nil {
		panic(err)
	}
	return r
}

func newResponse(result any, options ...jsonrpc.ResponseOption) *jsonrpc.Response {
	resp, err := jsonrpc.NewResponse(result, options...)
	if err != nil {
		panic(err)
	}
	return resp
}
