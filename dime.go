package dime

// ByteService defines a interface for underline systems which want to communicate like
// in a flow based paradigm through the use of channels. It allows different services
// to create both adapters by which data is supplied and responded.
type ByteService interface {
	// Receive will return a channel which will allow you to continue reading from it till it is closed.
	Receive(string) (chan []byte, error)

	// Send will take the giving channel reading the provided data of that channel till the channel is closed
	// to signal the end of the transmission.
	Send(string, chan []byte) error

	// Done defines a signal to other pending services to know whether the service is still servicing
	// request.
	Done() chan struct{}

	// Service defines a function to be called to stop the service internal operation and to close
	// all read/write operations.
	Stop() error
}
