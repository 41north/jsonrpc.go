package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var responseTestCases = []struct {
	value Response
	json  string
}{
	{
		newResponse("hello"),
		"{\"result\":\"hello\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse("world", ResponseNumericId(1456)),
		"{\"id\":1456,\"result\":\"world\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse("foo", ResponseStringId("resp-1")),
		"{\"id\":\"resp-1\",\"result\":\"foo\",\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponse([]string{"hello", "world"}, ResponseNumericId(1456), ResponseVersion("1.0")),
		"{\"id\":1456,\"result\":[\"hello\",\"world\"],\"jsonrpc\":\"1.0\"}",
	},
	{
		newResponseError(Error{Code: 123, Message: "some error"}),
		"{\"error\":{\"code\":123,\"message\":\"some error\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(Error{Code: 3421, Message: "another error"}, ResponseNumericId(1456)),
		"{\"id\":1456,\"error\":{\"code\":3421,\"message\":\"another error\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(Error{Code: -123, Message: "some bug"}, ResponseStringId("resp-1")),
		"{\"id\":\"resp-1\",\"error\":{\"code\":-123,\"message\":\"some bug\"},\"jsonrpc\":\"2.0\"}",
	},
	{
		newResponseError(Error{Code: 552, Message: "sort your code"}, ResponseNumericId(1456), ResponseVersion("1.0")),
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
		r, err := ResponseFromJSON([]byte(tc.json))
		assert.Nil(t, err, "failed to unmarshal from json")
		assert.Equal(t, tc.value, r)
	}
}
