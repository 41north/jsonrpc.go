package jsonrpc

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/41north/go-async"
	"github.com/juju/errors"
	gonanoid "github.com/matoous/go-nanoid"
	log "github.com/sirupsen/logrus"
)

var (
	idGen = func() string { return gonanoid.MustID(20) }

	ErrClosed = errors.ConstError("connection has been closed")
)

type (
	ResponseFuture = async.Future[async.Result[*Response]]
	RequestHandler = func(req Request)
	CloseHandler   = func(err error)
)

type Client interface {
	Connect() error

	Send(req Request, resp *Response) error
	SendContext(ctx context.Context, req Request, resp *Response) error
	SendAsync(req Request) ResponseFuture

	SetCloseHandler(handler CloseHandler)
	SetRequestHandler(handler RequestHandler)

	Close() error
}

type client struct {
	dialer       Dialer
	conn         Connection
	inFlight     sync.Map
	log          *log.Entry
	closed       atomic.Bool
	reqHandler   RequestHandler
	closeError   error
	closeHandler CloseHandler
}

func NewClient(dialer Dialer) Client {
	return &client{
		dialer: dialer,
	}
}

func (c *client) Connect() error {
	conn, err := c.dialer.Dial()
	if err != nil {
		return err
	}

	c.conn = conn
	c.inFlight = sync.Map{}
	c.log = log.WithField("connectionId", "tbd")

	go c.readMessages()

	return nil
}

func (c *client) SetRequestHandler(handler RequestHandler) {
	c.reqHandler = handler
}

func (c *client) SetCloseHandler(handler CloseHandler) {
	c.closeHandler = handler
}

func (c *client) readMessages() {
	for !c.closed.Load() {
		// read the next response
		bytes, err := c.conn.Read()
		if err != nil {
			// set the client has closed and break out of the read loop
			if err == ErrClosed {
				c.closeError = err
				c.Close()
				break
			}

			// otherwise log the error
			c.log.WithError(err).Error("read failure")
		}

		hasMethod := strings.Contains(string(bytes), "method")
		if hasMethod {
			// we assume this is a notification
			var req Request
			if err := json.Unmarshal(bytes, &req); err != nil {
				c.log.WithError(err).Error("unmarshal failure")
			} else {
				c.reqHandler(req)
			}
		} else {
			// otherwise we assume it is a response
			var resp Response
			if err := json.Unmarshal(bytes, &resp); err != nil {
				c.log.WithError(err).Error("unmarshal failure")
			} else {
				c.onResponse(&resp)
			}
		}
	}
}

func (c *client) onResponse(resp *Response) {
	future, ok := c.inFlight.LoadAndDelete(string(resp.Id))
	if !ok {
		c.log.
			WithField("id", resp.Id).
			Warn("response received with unrecognised id")
	}
	future.(ResponseFuture).Set(async.NewResult[*Response](resp))
}

func (c *client) Close() error {
	if c.closed.CompareAndSwap(false, true) {
		// cancel any in flight requests
		c.inFlight.Range(func(key, value any) bool {
			value.(ResponseFuture).Set(async.NewResultErr[*Response](ErrClosed))
			return true
		})

		if c.closeHandler != nil {
			c.closeHandler(c.closeError)
		}

		return nil
	} else {
		return ErrClosed
	}
}

func (c *client) Send(req Request, resp *Response) error {
	return c.SendContext(context.Background(), req, resp)
}

func (c *client) SendContext(ctx context.Context, req Request, resp *Response) error {
	future := c.SendAsync(req)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-future.Get():
		r, err := result.Unwrap()
		if err != nil {
			return err
		}
		// TODO can this copy be removed?
		resp.Id = r.Id
		resp.Result = r.Result
		resp.Error = r.Error
		resp.Version = r.Version
		return nil
	}
}

func (c *client) SendAsync(req Request) ResponseFuture {
	// create a future for returning the result
	future := async.NewFuture[async.Result[*Response]]()

	// ensure a request id
	if err := req.EnsureId(idGen); err != nil {
		future.Set(async.NewResultErr[*Response](err))
		return future
	}

	if c.closed.Load() {
		// short circuit
		future.Set(async.NewResultErr[*Response](ErrClosed))
		return future
	}

	// marshal to json
	bytes, err := json.Marshal(req)
	if err != nil {
		future.Set(async.NewResultErr[*Response](errors.Annotate(err, "failed to marshal request to json")))
		return future
	}

	// create an in flight entry
	c.inFlight.Store(string(req.Id), future)

	// send the request
	if err := c.conn.Write(bytes); err != nil {
		future.Set(async.NewResultErr[*Response](err))
	}

	return future
}
