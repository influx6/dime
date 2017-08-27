Dime
--------
Dime provides streaming structures for native go types using channels, which allows services, message queues
and other systems to be exposed to match such interfaces easily.

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


## A Dime Adapter
An adapter in Dime, is simply any function that can take a channel of a type and return a channel of another type, where by it does certain transformation within to produce such a conversion. As an example the [BoolService](./services/bool_service.go) adapters.


```go
// BoolFromByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolFromByteAdapterWithContext func(CancelContext, chan []byte) chan bool

// BoolToByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolToByteAdapter func(CancelContext, chan bool) chan []byte
```

## A Dime Service
A service in Dime, is simply any implementation that match a giving interface type, such has the [ByteService](./services/byte_service.go).

_Dime contains auto-generated code for all go's base types supported for usage outside of the package in [Services](./services) directory._

```go
type MonoBoolService interface {
	ReadErrors() <-chan error
	Read() (<-chan bool, error)
	Write(<-chan bool) error
	Done() chan struct{}
	Stop() error
}
```

```go
type BoolService interface {
	ReadErrors() <-chan error
	Read(string) (<-chan bool, error)
	Write(string, <-chan bool) error
	Done() chan struct{}
	Stop() error
}
```

