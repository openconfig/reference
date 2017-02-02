package fake

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/openconfig/reference/telemetry/pathtree"
)

func TestStream(t *testing.T) {
	sleep = func(time.Duration) {
		return
	}
	events := func() []Event {
		e := []Event{}
		for i := 0; i < 1000; i++ {
			e = append(e, Event{
				Value: &pathtree.PathVal{
					Path: pathtree.Path{"a", "b", fmt.Sprintf("%d", i)},
					Val:  i,
				},
			})
		}
		return e
	}()
	tests := []struct {
		desc       string
		e          []Event
		err        bool
		long       bool
		quickClose bool
	}{{
		desc: "No Events",
		err:  true,
	}, {
		desc: "Empty Events",
		e:    []Event{},
		err:  true,
	}, {
		desc: "Single Value",
		e: []Event{{
			Value: &pathtree.PathVal{
				Path: pathtree.Path{"a", "b"},
				Val:  5,
			},
		}},
	}, {
		desc: "Multi Value",
		e:    events,
	}, {
		desc: "Multi Value - long read",
		e:    events,
		long: true,
	}, {
		desc:       "Multi Value - quick close",
		e:          events,
		quickClose: true,
	}, {
		desc: "Error Value",
		e: []Event{{
			Value: SubscribeError{Msg: "foo"},
		}},
	}, {
		desc: "Invalid Value",
		e: []Event{{
			Value: 5,
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Log(tt.desc)
			s := NewStream(tt.e)
			if s == nil {
				if !tt.err {
					t.Fatalf("NewStream(%v) unexpected nil", tt.e)
				}
				return
			}
			readLen := len(tt.e)
			if tt.long {
				readLen++
			}
			if tt.quickClose {
				s.mu.Lock()
				close(s.closed)
				s.mu.Unlock()
				<-s.e
				return
			}
			for i := 0; i < readLen; i++ {
				v, err := s.Next()
				switch err.(type) {
				default:
					if err == io.EOF {
						if tt.long {
							continue
						}
						t.Fatalf("Next() with long read failed want io.EOF")
					}
					if !reflect.DeepEqual(tt.e[i].Value, err) {
						t.Fatalf("Next() failed: got %v, want %v", err, tt.e[i].Value)
					}
				case nil:
					if !reflect.DeepEqual(tt.e[i].Value, v) {
						t.Fatalf("Next() failed: got %v, want %v", v, tt.e[i].Value)
					}
				case StreamError: // this is caused by invalid value.
				}
			}
			s.Close()
		})
	}
}

func TestClient(t *testing.T) {
	tests := []struct {
		desc string
		in   []Event
		sub  pathtree.Path
		err  error
	}{{
		desc: "Empty events",
		err:  SubscribeError{Msg: "SubscribeError"},
	}, {
		desc: "Single events",
		in:   []Event{{Value: &pathtree.PathVal{Val: 1}}},
	}}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c := New(tt.in)
			s, err := c.Subscribe(tt.sub)
			if err != nil {
				if tt.err != err {
					t.Fatalf("Subscribe() unexpected error: got %v, want %v", err, tt.err)
				}
				return
			}
			e, err := s.Next()
			if err != nil {
				t.Fatalf("Next() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(tt.in[0].Value, e) {
				t.Fatalf("Next() failed: got %v, want %v", tt.in[0].Value, e)
			}
			s.Close()
		})
	}
}
