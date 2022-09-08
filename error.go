package jsonrpc

import (
	"encoding/json"
	"fmt"
)

var (
	ErrInvalidRequest = Error{
		Code:    -32600,
		Message: "invalid request",
	}
	ErrMethodNotFound = Error{
		Code:    -32601,
		Message: "method not found",
	}
	ErrInvalidParams = Error{
		Code:    -32602,
		Message: "invalid params",
	}
	ErrInternal = Error{
		Code:    -32603,
		Message: "internal error",
	}
	ErrParse = Error{
		Code:    -32700,
		Message: "parse error",
	}
)

type Error struct {
	Code    int32           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error renders e to a human-readable string for the error interface.
func (e Error) Error() string { return fmt.Sprintf("[%d] %s", e.Code, e.Message) }
