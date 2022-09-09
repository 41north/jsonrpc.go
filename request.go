package jsonrpc

import (
	"encoding/json"

	"github.com/juju/errors"
)

type IntOrString interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		string
}

func RequestVersion(version string) RequestOption {
	return func(opts *RequestOptions) error {
		opts.Version = version
		return nil
	}
}

func RequestStringId(id string) RequestOption {
	return func(opts *RequestOptions) error {
		opts.Id = id
		return nil
	}
}

func RequestNumericId(id int) RequestOption {
	return func(opts *RequestOptions) error {
		opts.Id = int64(id)
		return nil
	}
}

type RequestOption = func(opts *RequestOptions) error

type RequestOptions struct {
	Version string
	Id      any
}

func DefaultRequestOptions() RequestOptions {
	return RequestOptions{
		Version: "2.0",
	}
}

// newRequest is an internal test utility for creating request objects without having to handle
// the possible error, panicking instead.
func newRequest(method string, params any, options ...RequestOption) Request {
	r, err := NewRequest(method, params, options...)
	if err != nil {
		panic(err)
	}
	return r
}

func NewRequest(method string, params any, options ...RequestOption) (Request, error) {
	opts := DefaultRequestOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	var err error
	var paramBytes json.RawMessage
	if params != nil {
		paramBytes, err = json.Marshal(params)
		if err != nil {
			return nil, errors.New("failed to marshal params to json")
		}
	}

	return &request{id: opts.Id, method: method, params: paramBytes, jsonRpc: opts.Version}, nil
}

type Request interface {
	Id() any
	Method() string
	Params() json.RawMessage
	JsonRpc() string
	EnsureId(generator func() string)
}

type request struct {
	id      any
	method  string
	params  json.RawMessage
	jsonRpc string
}

func (r *request) Id() any {
	return r.id
}

func (r *request) Method() string {
	return r.method
}

func (r *request) Params() json.RawMessage {
	return r.params
}

func (r *request) JsonRpc() string {
	return r.jsonRpc
}

func (r *request) EnsureId(generator func() string) {
	if r.id == nil {
		r.id = generator()
	}
}

type requestDto struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	JsonRpc string          `json:"jsonrpc,omitempty"`
}

func (r *request) MarshalJSON() ([]byte, error) {
	var err error
	var id json.RawMessage

	if r.Id() != nil {
		id, err = json.Marshal(r.Id())
		if err != nil {
			return nil, errors.Annotate(err, "failed to marshal id to json")
		}
	}

	return json.Marshal(&requestDto{
		Id:      id,
		Method:  r.Method(),
		Params:  r.Params(),
		JsonRpc: r.JsonRpc(),
	})
}

func UnmarshalId(data json.RawMessage) (any, error) {
	// nil check first
	if data == nil {
		return nil, nil
	}

	var id any
	if err := json.Unmarshal(data, &id); err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal id")
	}

	switch v := id.(type) {
	case float64:
		// numeric values are decoded by json.Unmarshal as float64
		// we need to coerce into int64
		id = int64(v)

	case string:
		// do nothing
	default:
		return nil, errors.New("invalid id type, expected int64 or string")
	}

	return id, nil
}

func RequestFromJSON(data []byte) (Request, error) {
	var dto requestDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	id, err := UnmarshalId(dto.Id)
	if err != nil {
		return nil, err
	}

	return &request{
		id:      id,
		method:  dto.Method,
		params:  dto.Params,
		jsonRpc: dto.JsonRpc,
	}, nil
}
