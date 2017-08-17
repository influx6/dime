Dime
--------
Dime provides streaming structures for native go types using channels, which allows services, message queues
and other systems to be exposed to match such interfaces easily, more so, allowing the creation of adapters
that transform existing data types into others.

_Dime is inspired from the work done on [Vice](https://github.com/matryer/vice) by [Mat Ryer](https://github.com/matryer)._

_Dime uses code generation to create all the types interface and adapters using the [Moz](https://github.com/influx6/moz)._


## Custom Interface And Adapters
Dime provides easy means of creating new service types and adapters by simply declaring specific annotations in the package comments which indicate the template and details necessary to generate those artifacts.

For example, within the [dime.go](./doc.go) package comments, you would see comment lines like below:

```go
// @templaterTypesFor(id => Service, filename => bool_service.go, ServiceName => BoolService, Type => bool)
// @templaterTypesFor(id => Service, filename => bool_slice_service.go, ServiceName => BoolSliceService, Type => []bool)
```

These commentary above is responsible for the generation of the [BoolService](./bool_service.go) and [BoolSliceService](./bool_slice_service.go).


Where each `@templateTypesFor` dictate the giving details needed to use the default dime template to create new code specific for different internal go types. These easily can be customized to create new interfaces and adapters for specific custom structs as exposed services.

_To learn more about annotation code generation, see [Moz](https://github.com/influx6/moz)._


## A Dime Service
A service in Dime, is simply any implementation that match a giving interface type, such has the [ByteService](./services/byte_service.go).

```go
type ByteService interface {
	// Receive will return a channel which will allow reading from the Service it till it is closed.
	Receive(string) (chan byte, error)

	// Send will take the a channel, which will be written into the Service for it's internal processing
	// and the Service will continue to read form the channel till it is closed.
	// Useful for collating/collecting services.
	Send(string, chan byte) error

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
```

## A Dime Service Adapter
A adapter in Dime, is simply any function that can take a channel of a type and return a channel of another type, where by it does certain transformation within to produce such a conversion. As an example the [BoolService](./services/bool_service.go) adapters.


```go
// BoolServiceByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolServiceByteAdapter func(chan []byte) chan bool

// BoolServiceToByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolServiceToByteAdapter func(chan bool) chan []byte

// BoolService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve bool values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type BoolService interface {
	// Receive will return a channel which will allow reading from the Service it till it is closed.
	Receive(string) (chan bool, error)

	// Send will take the a channel, which will be written into the Service for it's internal processing
	// and the Service will continue to read form the channel till it is closed.
	// Useful for collating/collecting services.
	Send(string, chan bool) error

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
```