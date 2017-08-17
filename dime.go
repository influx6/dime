// Package dime implements a Service system for creating underline streaming backends
// exposed through channels and adapters. The types and structures are generated from
// the interface below.
/* All types of services and adapters are generated from the template below.
 @templater(id => Service, gen => Partial.Go, {

 	//go:generate moz generate-file -fromFile ./{{sel "filename"}} -toDir ./impl/{{sel "ServiceName" | lower}}

	{{ $noadapter := sel "NoAdapter" }}
	{{ if eq $noadapter "" }}
	// {{ sel "ServiceName" }}ByteAdapter defines a function that that will take a channel of bytes and return a channel of {{ sel "Type"}}.
 	type {{ sel "ServiceName" }}ByteAdapter func(chan []byte) chan {{sel "Type"}}

	// {{ sel "ServiceName" }}ToByteAdapter defines a function that that will take a channel of bytes and return a channel of {{ sel "Type"}}.
 	type {{ sel "ServiceName" }}ToByteAdapter func(chan {{sel "Type"}}) chan []byte
	{{ end}}

	// {{ sel "ServiceName" }} defines a interface for underline systems which want to communicate like
	// in a stream using channels to send/recieve {{sel "Type"}} values. It allows different services to create adapters to
	// transform data coming in and out from the Service.
	// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
	// @iface
 	type {{ sel "ServiceName" }} interface {
		// Receive will return a channel which will allow reading from the Service it till it is closed.
		Receive(string) (<-chan {{ sel "Type"}}, error)

		// Send will take the a channel, which will be written into the Service for it's internal processing
		// and the Service will continue to read form the channel till the channel is closed.
		// Useful for collating/collecting services.
		Send(string, chan<- {{sel "Type"}}) error

		// Done defines a signal to other pending services to know whether the Service is still servicing
		// request.
		Done() chan struct{}

		// Errors returns a channel which signals services to know whether the Service is still servicing
		// request.
		Errors() (chan error)

		// Service defines a function to be called to stop the Service internal operation and to close
		// all read/write operations.
		Stop() error
	}

})
*/
//
// Below are types annotations used to generate different Service types based on a giving type.
//
// @templaterTypesFor(id => Service, filename => map_service.go, ServiceName => MapService, Type => map[string]string)
// @templaterTypesFor(id => Service, filename => map_slice_service.go, ServiceName => MapSliceService, Type => []map[string]string)
//
// @templaterTypesFor(id => Service, filename => map_of_any_service.go, ServiceName => MapOfAnyService, Type => map[interface{}]interface{})
// @templaterTypesFor(id => Service, filename => map_of_any_slice_service.go, ServiceName => MapOfAnySliceService, Type => []map[interface{}]interface{})
//
// @templaterTypesFor(id => Service, filename => struct_service.go, ServiceName => StructService, Type => struct{})
// @templaterTypesFor(id => Service, filename => interface_service.go, ServiceName => InterfaceService, Type => interface{})
// @templaterTypesFor(id => Service, filename => interface_slice_service.go, ServiceName => InterfaceSliceService, Type => []interface{})
//
// @templaterTypesFor(id => Service, filename => string_service.go, ServiceName => StringService, Type => string)
// @templaterTypesFor(id => Service, filename => string_slice_service.go, ServiceName => StringSliceService, Type => []string)
//
// @templaterTypesFor(id => Service, filename => error_service.go, ServiceName => ErrorService, Type => error)
// @templaterTypesFor(id => Service, filename => error_slice_service.go, ServiceName => ErrorSliceService, Type => []error)
//
// @templaterTypesFor(id => Service, filename => bool_service.go, ServiceName => BoolService, Type => bool)
// @templaterTypesFor(id => Service, filename => bool_slice_service.go, ServiceName => BoolSliceService, Type => []bool)
//
// @templaterTypesFor(id => Service, NoAdapter => true, filename => byte_service.go, ServiceName => ByteService, Type => byte)
// @templaterTypesFor(id => Service, NoAdapter => true, filename => rune_service.go, ServiceName => RuneService, Type => rune)
// @templaterTypesFor(id => Service, NoAdapter => true, filename => bytes_service.go, ServiceName => BytesService, Type => []byte)
//
// @templaterTypesFor(id => Service, filename => float32_slice_service.go, ServiceName => Float32SliceService, Type => []float32)
// @templaterTypesFor(id => Service, filename => float64_slice_service.go, ServiceName => Float64SliceService, Type => float64)
// @templaterTypesFor(id => Service, filename => float32_service.go, ServiceName => Float32Service, Type => float32)
// @templaterTypesFor(id => Service, filename => float64_service.go, ServiceName => Float64Service, Type => float64)
//
// @templaterTypesFor(id => Service, filename => complex64_service.go, ServiceName => Complex64Service, Type => complex64)
// @templaterTypesFor(id => Service, filename => complex128_service.go, ServiceName => Complex128Service, Type => complex128)
// @templaterTypesFor(id => Service, filename => complex64_slice_service.go, ServiceName => Complex64SliceService, Type => complex64)
// @templaterTypesFor(id => Service, filename => complex128_slice_service.go, ServiceName => Complex128SliceService, Type => complex128)
//
// @templaterTypesFor(id => Service, filename => int_slice_service.go, ServiceName => IntSliceService, Type => []int)
// @templaterTypesFor(id => Service, filename => int8_slice_service.go, ServiceName => Int8SliceService, Type => []int8)
// @templaterTypesFor(id => Service, filename => int16_slice_service.go, ServiceName => Int16SliceService, Type => []int16)
// @templaterTypesFor(id => Service, filename => int32_slice_service.go, ServiceName => Int32SliceService, Type => []int32)
// @templaterTypesFor(id => Service, filename => int64_slice_service.go, ServiceName => Int64SliceService, Type => []int64)
//
// @templaterTypesFor(id => Service, filename => int_service.go, ServiceName => IntService, Type => int)
// @templaterTypesFor(id => Service, filename => int8_service.go, ServiceName => Int8Service, Type => int8)
// @templaterTypesFor(id => Service, filename => int16_service.go, ServiceName => Int16Service, Type => int16)
// @templaterTypesFor(id => Service, filename => int32_service.go, ServiceName => Int32Service, Type => int32)
// @templaterTypesFor(id => Service, filename => int64_service.go, ServiceName => Int64Service, Type => int64)
//
// @templaterTypesFor(id => Service, filename => uint_service.go, ServiceName => UIntService, Type => uint)
// @templaterTypesFor(id => Service, filename => uint8_service.go, ServiceName => UInt8Service, Type => uint8)
// @templaterTypesFor(id => Service, filename => uint32_service.go, ServiceName => UInt32Service, Type => uint32)
// @templaterTypesFor(id => Service, filename => uint64_service.go, ServiceName => UInt64Service, Type => uint64)
//
// @templaterTypesFor(id => Service, filename => uint_slice_service.go, ServiceName => UIntSliceService, Type => []uint)
// @templaterTypesFor(id => Service, filename => uint8_slice_service.go, ServiceName => UInt8SliceService, Type => []uint8)
// @templaterTypesFor(id => Service, filename => uint32_slice_service.go, ServiceName => UInt32SliceService, Type => []uint32)
// @templaterTypesFor(id => Service, filename => uint64_slice_service.go, ServiceName => UInt64SliceService, Type => []uint64)
//
package dime
