package jsonrpc

import (
	"context"
)

type Connection interface {
	Write(req Request) error
	Read() (Response, error)
	Close() error
}

type Dialer interface {
	Dial() (Connection, error)
	DialContext(ctx context.Context) (Connection, error)
}
