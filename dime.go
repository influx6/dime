// Package dime implements a service system for creating underline streaming backends
// exposed through channels and adapters. The types and structures are generated from
// the interface below.
/* All types of services and adapters are generated from the template below.
 @templater(id => Service, gen => Partial.Go, {

	// {{ sel "ServiceName" }} defines a interface for underline systems which want to communicate like
	// in a stream using channels to send/recieve {{sel "Type"}} values. It allows different services to create adapters to
	// transform data coming in and out from the service.
	// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
 	type {{ sel "ServiceName" }} interface {
		// Receive will return a channel which will allow reading from the service it till it is closed.
		Receive(string) (chan {{ sel "Type"}}, error)

		// Send will take the a channel, which will be written into the service for it's internal processing
		// and the service will continue to read form the channel till it is closed.
		// Useful for collating/collecting services.
		Send(string, chan {{sel "Type"}}) error

		// Done defines a signal to other pending services to know whether the service is still servicing
		// request.
		Done() chan struct{}

		// Errors returns a channel which signals services to know whether the service is still servicing
		// request.
		Errors() (chan error)

		// Service defines a function to be called to stop the service internal operation and to close
		// all read/write operations.
		Stop() error
	}

})
*/
//
// Below are types annotations used to generate different service types based on a giving type.
//
// @templateTypesFor(id => Service, filename => map_service.go, ServiceName => "MapService", Type => map[string]string)
// @templateTypesFor(id => Service, filename => map_slice_service.go, ServiceName => "MapSliceService", Type => []map[string]string)
//
// @templateTypesFor(id => Service, filename => map_of_any_service.go, ServiceName => "MapOfAnyService", Type => map[interface{}]interface{})
// @templateTypesFor(id => Service, filename => map_of_any_slice_service.go, ServiceName => "MapOfAnySliceService", Type => []map[interface{}]interface{})
//
// @templateTypesFor(id => Service, filename => struct_service.go, ServiceName => "StructService", Type => struct{})
// @templateTypesFor(id => Service, filename => interface_service.go, ServiceName => "InterfaceService", Type => interface{})
// @templateTypesFor(id => Service, filename => interface_slice_service.go, ServiceName => "InterfaceSliceService", Type => []interface{})
//
// @templateTypesFor(id => Service, filename => string_service.go, ServiceName => "StringService", Type => string)
// @templateTypesFor(id => Service, filename => string_slice_service.go, ServiceName => "StringService", Type => []string)
//
// @templateTypesFor(id => Service, filename => error_service.go, ServiceName => "ErrorService", Type => error)
// @templateTypesFor(id => Service, filename => error_slice_service.go, ServiceName => "ErrorSliceService", Type => []error)
//
// @templateTypesFor(id => Service, filename => bool_service.go, ServiceName => "BoolService", Type => bool)
// @templateTypesFor(id => Service, filename => bool_slice_service.go, ServiceName => "BoolSliceService", Type => []bool)
//
// @templateTypesFor(id => Service, filename => byte_service.go, ServiceName => "ByteService", Type => byte)
// @templateTypesFor(id => Service, filename => rune_service.go, ServiceName => "RuneService", Type => rune)
// @templateTypesFor(id => Service, filename => bytes_service.go, ServiceName => "BytesService", Type => []byte)
//
// @templateTypesFor(id => Service, filename => float32_slice_service.go, ServiceName => "Float32SliceService", Type => []float32)
// @templateTypesFor(id => Service, filename => float64_slice_service.go, ServiceName => "Float64SliceService", Type => float64)
// @templateTypesFor(id => Service, filename => float32_service.go, ServiceName => "Float32Service", Type => float32)
// @templateTypesFor(id => Service, filename => float64_service.go, ServiceName => "Float64Service", Type => float64)
//
// @templateTypesFor(id => Service, filename => complex64_service.go, ServiceName => "Complex32Service", Type => complex64)
// @templateTypesFor(id => Service, filename => complex128_service.go, ServiceName => "Complex64Service", Type => complex128)
// @templateTypesFor(id => Service, filename => complex64_slice_service.go, ServiceName => "Complex32SliceService", Type => complex64)
// @templateTypesFor(id => Service, filename => complex128_slice_service.go, ServiceName => "Complex64SliceService", Type => complex128)
//
// @templateTypesFor(id => Service, filename => int_slice_service.go, ServiceName => "IntSliceService", Type => []int)
// @templateTypesFor(id => Service, filename => int8_slice_service.go, ServiceName => "Int8SliceService", Type => []int8)
// @templatetypesfor(id => service, filename => int16_slice_service.go, servicename => "Int16SliceService", type => []int16)
// @templatetypesfor(id => service, filename => int32_slice_service.go, servicename => "Int32SliceService", type => []int32)
// @templatetypesfor(id => service, filename => int64_slice_service.go, servicename => "Int64SliceService", type => []int64)
//
// @templateTypesFor(id => Service, filename => int_service.go, ServiceName => "IntService", Type => int)
// @templateTypesFor(id => Service, filename => int8_service.go, ServiceName => "Int8Service", Type => int8)
// @templatetypesfor(id => service, filename => int16_service.go, servicename => "Int16Service", type => int16)
// @templatetypesfor(id => service, filename => int32_service.go, servicename => "Int32Service", type => int32)
// @templatetypesfor(id => service, filename => int64_service.go, servicename => "Int54Service", type => int64)
//
// @templateTypesFor(id => Service, filename => uint_service.go, ServiceName => "UIntService", Type => uint)
// @templateTypesFor(id => Service, filename => uint8_service.go, ServiceName => "UInt8Service", Type => uint8)
// @templatetypesfor(id => service, filename => uint32_service.go, servicename => "UInt32Service", type => uint32)
// @templatetypesfor(id => service, filename => uint64_service.go, servicename => "UInt54Service", type => uint64)
//
// @templateTypesFor(id => Service, filename => uint_slice_service.go, ServiceName => "UIntSliceService", Type => []uint)
// @templateTypesFor(id => Service, filename => uint8_slice_service.go, ServiceName => "UInt8SliceService", Type => []uint8)
// @templatetypesfor(id => service, filename => uint32_slice_service.go, servicename => "UInt32SliceService", type => []uint32)
// @templatetypesfor(id => service, filename => uint64_slice_service.go, servicename => "UInt64SliceService", type => []uint64)
//
package dime
