package jsonrpc

import "encoding/json"

type IntOrString interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		string
}

type Request struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	JsonRpc string          `json:"jsonrpc"`
}

func (r *Request) UnmarshalId() (any, error) {
	return unmarshalId(r.Id)
}

func (r *Request) WithStringId(id string) error {
	bytes, err := json.Marshal(id)
	r.Id = bytes
	return err
}

func (r *Request) WithIntegerId(id uint64) error {
	bytes, err := json.Marshal(id)
	r.Id = bytes
	return err
}
