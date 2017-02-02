package fake

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/openconfig/reference/telemetry/pathtree"
)

var (
	sleep = time.Sleep
)

type Event struct {
	Value     interface{}
	Timestamp time.Time
}

type Stream struct {
	mu       sync.Mutex
	Events   []Event
	e        chan Event
	closed   chan struct{}
	finalize chan struct{}
}

func NewStream(e []Event) *Stream {
	if len(e) == 0 {
		return nil
	}
	s := &Stream{
		Events: e,
		e:      make(chan Event),
		closed: make(chan struct{}),
	}
	// start processing
	go func() {
		t := time.Unix(0, 0)
		for _, e := range s.Events {
			d := t.Sub(e.Timestamp)
			if d > 0 {
				sleep(d)
			}
			select {
			case <-s.closed:
				close(s.e)
				return
			case s.e <- e:
			}
		}
		close(s.e)
	}()
	return s
}

func (s *Stream) Next() (*pathtree.PathVal, error) {
	e, ok := <-s.e
	if !ok {
		s.mu.Lock()
		if s.closed != nil {
			close(s.closed)
			s.closed = nil
		}
		s.mu.Unlock()
		return nil, io.EOF
	}
	switch v := e.Value.(type) {
	default:
		return nil, StreamError{Msg: fmt.Sprintf("invalid Event type: %T", e)}
	case error:
		return nil, v
	case *pathtree.PathVal:
		return v, nil
	}
}

func (s *Stream) Close() {
	s.mu.Lock()
	if s.closed != nil {
		close(s.closed)
		s.closed = nil
	}
	s.mu.Unlock()
	<-s.e
}

type Client struct {
	mu sync.Mutex
	s  *Stream
}

func New(e []Event) *Client {
	return &Client{
		s: NewStream(e),
	}
}

func (c *Client) Get(p pathtree.Path) (*pathtree.Branch, error) {
	return nil, GetError{Msg: "GetError"}
}

func (c *Client) Set(p pathtree.Path, b *pathtree.Branch) (*pathtree.Branch, error) {
	return nil, SetError{Msg: "SetError"}
}

func (c *Client) Subscribe(p pathtree.Path) (*Stream, error) {
	if c.s == nil {
		return nil, SubscribeError{Msg: "SubscribeError"}
	}
	return c.s, nil
}

// Errors returned by underlying actions.

type GetError struct {
	Msg string
}

func (e GetError) Error() string {
	return e.Msg
}

type SetError struct {
	Msg string
}

func (e SetError) Error() string {
	return e.Msg
}

type SubscribeError struct {
	Msg string
}

func (e SubscribeError) Error() string {
	return e.Msg
}

type StreamError struct {
	Msg string
}

func (e StreamError) Error() string {
	return e.Msg
}
