package jsonrpc

import (
	"context"
)

type Connection interface {
	Write(data []byte) error
	Read() ([]byte, error)
	Close() error
}

type Dialer interface {
	Dial() (Connection, error)
	DialContext(ctx context.Context) (Connection, error)
}
