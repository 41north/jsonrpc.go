package jsonrpc

import (
	"encoding/json"

	"github.com/juju/errors"
)

func ResponseStringId(id string) ResponseOption {
	return func(opts *ResponseOptions) error {
		bytes, err := json.Marshal(id)
		if err != nil {
			return err
		}
		opts.Id = bytes
		return nil
	}
}

func ResponseNumericId(id int) ResponseOption {
	return func(opts *ResponseOptions) error {
		bytes, err := json.Marshal(id)
		if err != nil {
			return err
		}
		opts.Id = bytes
		return nil
	}
}

func ResponseVersion(version string) ResponseOption {
	return func(opts *ResponseOptions) error {
		opts.Version = version
		return nil
	}
}

type ResponseOption = func(opts *ResponseOptions) error

type ResponseOptions struct {
	Id      json.RawMessage
	Version string
}

func DefaultResponseOptions() ResponseOptions {
	return ResponseOptions{
		Version: "2.0",
	}
}

func newResponse(result any, options ...ResponseOption) *Response {
	resp, err := NewResponse(result, options...)
	if err != nil {
		panic(err)
	}
	return resp
}

func NewResponse(result any, options ...ResponseOption) (*Response, error) {
	opts := DefaultResponseOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal result to json")
	}

	return &Response{Id: opts.Id, Result: resultBytes, Error: nil, Version: opts.Version}, nil
}

func newResponseError(error Error, options ...ResponseOption) *Response {
	resp, err := NewResponseError(error, options...)
	if err != nil {
		panic(err)
	}
	return resp
}

func NewResponseError(error Error, options ...ResponseOption) (*Response, error) {
	opts := DefaultResponseOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}
	return &Response{Id: opts.Id, Result: nil, Error: &error, Version: opts.Version}, nil
}

type Response struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	Version string          `json:"jsonrpc"`
}

func (r *Response) UnmarshalId(payload any) error {
	return json.Unmarshal(r.Id, &payload)
}

func (r *Response) UnmarshalResult(payload any) error {
	if r.Error != nil {
		return r.Error
	}
	return json.Unmarshal(r.Result, &payload)
}
