// Autogenerated using the moz templater annotation.
//
//
package services

//go:generate moz generate-file -fromFile ./interface_service.go -toDir ./impl/interfaceservice

// InterfaceServiceByteAdapter defines a function that that will take a channel of bytes and return a channel of interface{}.
type InterfaceServiceByteAdapter func(chan []byte) chan interface{}

// InterfaceServiceToByteAdapter defines a function that that will take a channel of bytes and return a channel of interface{}.
type InterfaceServiceToByteAdapter func(chan interface{}) chan []byte

// InterfaceService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve interface{} values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type InterfaceService interface {
	// Receive will return a channel which will allow reading from the Service it till it is closed.
	Receive(string) (chan interface{}, error)

	// Send will take the a channel, which will be written into the Service for it's internal processing
	// and the Service will continue to read form the channel till it is closed.
	// Useful for collating/collecting services.
	Send(string, chan interface{}) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() chan struct{}

	// Errors returns a channel which signals services to know whether the Service is still servicing
	// request.
	Errors() chan error

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}