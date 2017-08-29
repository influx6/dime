package services

import "time"

var (
	// defaultSendWithBeforeAbort defines the time to await receving a message else aborting.
	defaultSendWithBeforeAbort = 3 * time.Second
)

// CancelContext defines a interface which returns a single method that returns a channel for cancellation signal.
type CancelContext interface {
	Done() <-chan struct{}
}

// ValueBagContext defines a interface which exposes a method for retrieval of a value
// through a string key.
type ValueBagContext interface {
	Get(string) (interface{}, error)
}
