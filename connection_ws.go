package jsonrpc

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type webSocketConnection struct {
	conn *websocket.Conn
}

func (w *webSocketConnection) Write(data []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *webSocketConnection) Read() ([]byte, error) {
	msgType, bytes, err := w.conn.ReadMessage()
	if err != nil {

		log.WithError(err).Error("read failure")

		switch err.(type) {
		case *websocket.CloseError:
			// re-map error
			return nil, ErrClosed
		default:
			return nil, err
		}
	}

	if msgType != websocket.TextMessage {
		return nil, errors.Errorf("expected text message type, received a writer for %v", msgType)
	}

	if err != nil {
		return nil, errors.Annotate(err, "failed to read message")
	}

	return bytes, nil
}

func (w *webSocketConnection) Close() error {
	return w.conn.Close()
}

type WebSocketDialer struct {
	Url           string
	RequestHeader http.Header
	// TODO expose more of the underlying ws options
}

func (w WebSocketDialer) Dial() (Connection, error) {
	return w.DialContext(context.Background())
}

func (w WebSocketDialer) DialContext(ctx context.Context) (Connection, error) {
	dialer := websocket.Dialer{}
	wsConn, _, err := dialer.DialContext(ctx, w.Url, w.RequestHeader)
	conn := webSocketConnection{conn: wsConn}
	return &conn, err
}
