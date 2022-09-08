package jsonrpc

import (
	"encoding/json"

	"github.com/juju/errors"
)

type Response struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	JsonRpc string          `json:"jsonrpc"`
}

func (r *Response) UnmarshalId() (any, error) {
	return unmarshalId(r.Id)
}

func (r *Response) UnmarshalResult(payload any) error {
	if r.Error != nil {
		return r.Error
	}
	return json.Unmarshal(r.Result, payload)
}

func unmarshalId(id json.RawMessage) (any, error) {
	var key any
	err := json.Unmarshal(id, &key)
	if err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal id")
	}
	switch key.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, string:
		return key, nil
	default:
		return nil, errors.Errorf("id field must be an integer or a string, found: %v", key)
	}
}
