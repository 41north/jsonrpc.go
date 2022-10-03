package jsonrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/41north/jsonrpc.go"

	"github.com/stretchr/testify/assert"
)

var responseTestCases = []struct {
	value *jsonrpc.Response
	json  string
}{
	{
		newResponse("hello"),
		"{\"result\":\"hello\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse("world", jsonrpc.ResponseNumericId(1456)),
		"{\"id\":1456,\"result\":\"world\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse("foo", jsonrpc.ResponseStringId("resp-1")),
		"{\"id\":\"resp-1\",\"result\":\"foo\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse([]string{"hello", "world"}, jsonrpc.ResponseNumericId(1456), jsonrpc.ResponseVersion("1.0")),
		"{\"id\":1456,\"result\":[\"hello\",\"world\"],\"jsonrpc\":\"1.0\"}",
	},
	{
		newResponseError(jsonrpc.Error{Code: 123, Message: "some error"}),
		"{\"error\":{\"code\":123,\"message\":\"some error\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(jsonrpc.Error{Code: 3421, Message: "another error"}, jsonrpc.ResponseNumericId(1456)),
		"{\"id\":1456,\"error\":{\"code\":3421,\"message\":\"another error\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(jsonrpc.Error{Code: -123, Message: "some bug"}, jsonrpc.ResponseStringId("resp-1")),
		"{\"id\":\"resp-1\",\"error\":{\"code\":-123,\"message\":\"some bug\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(jsonrpc.Error{Code: 552, Message: "sort your code"}, jsonrpc.ResponseNumericId(1456), jsonrpc.ResponseVersion("1.0")),
		"{\"id\":1456,\"error\":{\"code\":552,\"message\":\"sort your code\"},\"jsonrpc\":\"1.0\"}",
	},
}

func TestResponse_MarshalJSON(t *testing.T) {
	for _, tc := range responseTestCases {
		bytes, err := json.Marshal(tc.value)
		assert.Nil(t, err, "failed to marshal to json")
		assert.Equal(t, tc.json, string(bytes))
	}
}

func TestResponse_UnmarshalJSON(t *testing.T) {
	for _, tc := range responseTestCases {
		var resp jsonrpc.Response
		err := json.Unmarshal([]byte(tc.json), &resp)
		assert.Nil(t, err, "failed to unmarshal from json")
		assert.Equal(t, *tc.value, resp)
	}
}

func newResponseError(error jsonrpc.Error, options ...jsonrpc.ResponseOption) *jsonrpc.Response {
	resp, err := jsonrpc.NewResponseError(error, options...)
	if err != nil {
		panic(err)
	}
	return resp
}
