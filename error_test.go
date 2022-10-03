package jsonrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/41north/jsonrpc.go"

	"github.com/stretchr/testify/assert"
)

var dataBytes, _ = json.Marshal("Some data")

var errorTestCases = []struct {
	value any
	json  string
}{
	{jsonrpc.ErrInvalidRequest, "{\"code\":-32600,\"message\":\"invalid request\"}"},
	{jsonrpc.ErrMethodNotFound, "{\"code\":-32601,\"message\":\"method not found\"}"},
	{jsonrpc.ErrInvalidParams, "{\"code\":-32602,\"message\":\"invalid params\"}"},
	{jsonrpc.ErrInternal, "{\"code\":-32603,\"message\":\"internal error\"}"},
	{jsonrpc.ErrParse, "{\"code\":-32700,\"message\":\"parse error\"}"},
	{jsonrpc.Error{123, "error with data", dataBytes}, "{\"code\":123,\"message\":\"error with data\",\"data\":\"Some data\"}"},
}

func TestError_Marshal(t *testing.T) {
	for _, tt := range errorTestCases {
		bytes, err := json.Marshal(tt.value)
		assert.Nil(t, err, "failed to marshal to json")
		assert.Equal(t, tt.json, string(bytes))
	}
}

func TestError_Unmarshal(t *testing.T) {
	for _, tt := range errorTestCases {
		var e jsonrpc.Error
		err := json.Unmarshal([]byte(tt.json), &e)
		assert.Nil(t, err, "failed to unmarshal from json")
		assert.Equal(t, tt.value, e)
	}
}
