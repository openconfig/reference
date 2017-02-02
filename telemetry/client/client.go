// Package client provides the client interfaces for streaming telemetry.
// Since there are several versions of the backend implementations this
// abstracts the notification protocol to a simple stream of path values.
package client

import (
	"github.com/openconfig/reference/telemetry/pathtree"
)

// Stream provides an iterator for the events being streamed from the client.
type Stream interface {
	// Next iterates over notifications in the subscription.  Failure to call
	// Next() will block the client and backpressure the server.
	Next() (pathtree.Path, error)
	Close()
}

// Client provides a generic interface for interacting with underlying
// transports for streaming telemetry.
type Client interface {
	// Get gets the subtree at path.
	Get(pathtree.Path) (*pathtree.Branch, error)
	// Set sets the value at path.
	Set(pathtree.Path, *pathtree.Branch) error
	// Subscribe will subscribe to the path and return a stream
	// of pathtree.PathVal changes
	Subscribe(pathtree.Path) (Stream, error)
}
