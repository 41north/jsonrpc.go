package jsonrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/41north/jsonrpc.go"

	"github.com/stretchr/testify/assert"
)

var requestTestCases = []struct {
	value *jsonrpc.Request
	json  string
}{
	{
		newRequest("ping", "foo"),
		"{\"method\":\"ping\",\"params\":\"foo\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newRequest("pong", "bar", jsonrpc.RequestNumericId(1456)),
		"{\"id\":1456,\"method\":\"pong\",\"params\":\"bar\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newRequest("ping", nil, jsonrpc.RequestStringId("req-1")),
		"{\"id\":\"req-1\",\"method\":\"ping\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newRequest("pong", []string{"hello", "world"}, jsonrpc.RequestNumericId(554), jsonrpc.RequestVersion("1.0")),
		"{\"id\":554,\"method\":\"pong\",\"params\":[\"hello\",\"world\"],\"jsonrpc\":\"1.0\"}",
	},
}

func TestRequest_MarshalJSON(t *testing.T) {
	for _, tc := range requestTestCases {
		bytes, err := json.Marshal(tc.value)
		assert.Nil(t, err, "failed to marshal to json")
		assert.Equal(t, tc.json, string(bytes))
	}
}

func TestRequest_UnmarshalJSON(t *testing.T) {
	for _, tc := range requestTestCases {
		var req jsonrpc.Request
		err := json.Unmarshal([]byte(tc.json), &req)
		assert.Nil(t, err, "failed to unmarshal from json")
		assert.Equal(t, *tc.value, req)
	}
}

func TestRequest_UnmarshalParams(t *testing.T) {
	expected := []string{"hello", "world"}
	req := newRequest("ping", expected)
	var actual []string
	err := req.UnmarshalParams(&actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}
