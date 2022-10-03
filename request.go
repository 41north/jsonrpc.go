package jsonrpc

import (
	"encoding/json"

	"github.com/juju/errors"
)

func RequestVersion(version string) RequestOption {
	return func(opts *RequestOptions) error {
		opts.Version = version
		return nil
	}
}

func RequestStringId(id string) RequestOption {
	return func(opts *RequestOptions) error {
		bytes, err := json.Marshal(id)
		if err != nil {
			return err
		}
		opts.Id = bytes
		return nil
	}
}

func RequestNumericId(id int) RequestOption {
	return func(opts *RequestOptions) error {
		bytes, err := json.Marshal(id)
		if err != nil {
			return err
		}
		opts.Id = bytes
		return nil
	}
}

type RequestOption = func(opts *RequestOptions) error

type RequestOptions struct {
	Version string
	Id      json.RawMessage
}

func DefaultRequestOptions() RequestOptions {
	return RequestOptions{
		Version: "2.0",
	}
}

func NewRequest(method string, params any, options ...RequestOption) (*Request, error) {
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

	return &Request{Id: opts.Id, Method: method, Params: paramBytes, Version: opts.Version}, nil
}

type IdGenerator = func() string

type Request struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	Version string          `json:"jsonrpc,omitempty"`
}

func (r *Request) EnsureId(gen IdGenerator) error {
	if r.Id != nil {
		return nil
	}
	bytes, err := json.Marshal(gen())
	if err != nil {
		return err
	}
	r.Id = bytes
	return nil
}

func (r *Request) UnmarshalId(id any) error {
	return json.Unmarshal(r.Id, &id)
}

func (r *Request) UnmarshalParams(payload any) error {
	return json.Unmarshal(r.Params, &payload)
}
