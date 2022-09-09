package jsonrpc

import (
	"encoding/json"

	"github.com/juju/errors"
)

func ResponseStringId(id string) ResponseOption {
	return func(opts *ResponseOptions) error {
		opts.Id = id
		return nil
	}
}

func ResponseNumericId(id int) ResponseOption {
	return func(opts *ResponseOptions) error {
		opts.Id = int64(id)
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
	Id      any
	Version string
}

func DefaultResponseOptions() ResponseOptions {
	return ResponseOptions{
		Version: "2.0",
	}
}

func newResponse(result any, options ...ResponseOption) Response {
	resp, err := NewResponse(result, options...)
	if err != nil {
		panic(err)
	}
	return resp
}

func NewResponse(result any, options ...ResponseOption) (Response, error) {
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

	return &response{id: opts.Id, result: resultBytes, error: nil, jsonRpc: opts.Version}, nil
}

func newResponseError(error Error, options ...ResponseOption) Response {
	resp, err := NewResponseError(error, options...)
	if err != nil {
		panic(err)
	}
	return resp
}

func NewResponseError(error Error, options ...ResponseOption) (Response, error) {
	opts := DefaultResponseOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}
	return &response{id: opts.Id, result: nil, error: &error, jsonRpc: opts.Version}, nil
}

type Response interface {
	Id() any
	Result() json.RawMessage
	UnmarshalResult(payload any) error
	Error() *Error
	JsonRpc() string
}

type response struct {
	id      any
	result  json.RawMessage
	error   *Error
	jsonRpc string
}

func (r *response) Id() any {
	return r.id
}

func (r *response) Result() json.RawMessage {
	return r.result
}

func (r *response) Error() *Error {
	return r.error
}

func (r *response) JsonRpc() string {
	return r.jsonRpc
}

func (r *response) UnmarshalResult(payload any) error {
	if r.Error() != nil {
		return r.Error()
	}
	return json.Unmarshal(r.Result(), payload)
}

func (r *response) MarshalJSON() ([]byte, error) {
	var err error
	var id json.RawMessage

	if r.Id() != nil {
		id, err = json.Marshal(r.Id())
		if err != nil {
			return nil, errors.Annotate(err, "failed to marshal id to json")
		}
	}

	return json.Marshal(&responseDto{
		Id:      id,
		Result:  r.Result(),
		Error:   r.Error(),
		JsonRpc: r.JsonRpc(),
	})
}

func ResponseFromJSON(data []byte) (Response, error) {
	var dto responseDto
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	id, err := UnmarshalId(dto.Id)
	if err != nil {
		return nil, err
	}

	return &response{
		id:      id,
		result:  dto.Result,
		error:   dto.Error,
		jsonRpc: dto.JsonRpc,
	}, nil
}

type responseDto struct {
	Id      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	JsonRpc string          `json:"jsonrpc"`
}
