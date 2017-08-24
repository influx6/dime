// Package dime implements a Service system for creating underline streaming backends
// exposed through channels and adapters. The types and structures are generated from
// the interface below.
//
//
/*
@templater(id => Vars, gen => Partial.Go, {
		var (
			// defaultSendWithBeforeAbort defines the time to await receving a message else aborting.
			defaultSendWithBeforeAbort = 3 * time.Second
		)
})

@templater(id => Service, gen => Partial.Go, file => types.tml)
@templater(id => ServiceTest, gen => Partial.Go, file => types-slice-test.tml)
@templater(id => ServiceSliceTest, gen => Partial.Go, file => types-test.tml)

@templaterTypesFor(id => Vars, filename => vars.go)

Below are types annotations used to generate different Service types based on a giving type.

@templaterTypesFor(id => Service, filename => map_service.go, Name => Map, Type => map[string]string)
@templaterTypesFor(id => Service, filename => map_slice_service.go, Name => MapSlice, Type => []map[string]string)

@templaterTypesFor(id => Service, filename => map_of_any_service.go, Name => MapOfAny, Type => map[interface{}]interface{})
@templaterTypesFor(id => Service, filename => map_of_any_slice_service.go, Name => MapOfAnySlice, Type => []map[interface{}]interface{})

@templaterTypesFor(id => Service, filename => interface_service.go, Name => Interface, Type => interface{})
@templaterTypesFor(id => Service, filename => interface_slice_service.go, Name => InterfaceSlice, Type => []interface{})

@templaterTypesFor(id => Service, filename => string_service.go, Name => String, Type => string)
@templaterTypesFor(id => Service, filename => string_slice_service.go, Name => StringSlice, Type => []string)

@templaterTypesFor(id => Service, filename => error_service.go, Name => Error, Type => error)
@templaterTypesFor(id => Service, filename => error_slice_service.go, Name => ErrorSlice, Type => []error)

@templaterTypesFor(id => Service, filename => bool_service.go, Name => Bool, Type => bool)
@templaterTypesFor(id => Service, filename => bool_slice_service.go, Name => BoolSlice, Type => []bool)

@templaterTypesFor(id => Service, NoAdapter => true, filename => byte_service.go, Name => Byte, Type => byte)
@templaterTypesFor(id => Service, NoAdapter => true, filename => rune_service.go, Name => Rune, Type => rune)
@templaterTypesFor(id => Service, NoAdapter => true, filename => bytes_service.go, Name => Bytes, Type => []byte)

@templaterTypesFor(id => Service, filename => float32_slice_service.go, Name => Float32Slice, Type => []float32)
@templaterTypesFor(id => Service, filename => float64_slice_service.go, Name => Float64Slice, Type => float64)
@templaterTypesFor(id => Service, filename => float32_service.go, Name => Float32, Type => float32)
@templaterTypesFor(id => Service, filename => float64_service.go, Name => Float64, Type => float64)

@templaterTypesFor(id => Service, filename => complex64_service.go, Name => Complex64, Type => complex64)
@templaterTypesFor(id => Service, filename => complex128_service.go, Name => Complex128, Type => complex128)
@templaterTypesFor(id => Service, filename => complex64_slice_service.go, Name => Complex64Slice, Type => complex64)
@templaterTypesFor(id => Service, filename => complex128_slice_service.go, Name => Complex128Slice, Type => complex128)

@templaterTypesFor(id => Service, filename => int_slice_service.go, Name => IntSlice, Type => []int)
@templaterTypesFor(id => Service, filename => int8_slice_service.go, Name => Int8Slice, Type => []int8)
@templaterTypesFor(id => Service, filename => int16_slice_service.go, Name => Int16Slice, Type => []int16)
@templaterTypesFor(id => Service, filename => int32_slice_service.go, Name => Int32Slice, Type => []int32)
@templaterTypesFor(id => Service, filename => int64_slice_service.go, Name => Int64Slice, Type => []int64)

@templaterTypesFor(id => Service, filename => int_service.go, Name => Int, Type => int)
@templaterTypesFor(id => Service, filename => int8_service.go, Name => Int8, Type => int8)
@templaterTypesFor(id => Service, filename => int16_service.go, Name => Int16, Type => int16)
@templaterTypesFor(id => Service, filename => int32_service.go, Name => Int32, Type => int32)
@templaterTypesFor(id => Service, filename => int64_service.go, Name => Int64, Type => int64)

@templaterTypesFor(id => Service, filename => uint_service.go, Name => UInt, Type => uint)
@templaterTypesFor(id => Service, filename => uint8_service.go, Name => UInt8, Type => uint8)
@templaterTypesFor(id => Service, filename => uint32_service.go, Name => UInt32, Type => uint32)
@templaterTypesFor(id => Service, filename => uint64_service.go, Name => UInt64, Type => uint64)

@templaterTypesFor(id => Service, filename => uint_slice_service.go, Name => UIntSlice, Type => []uint)
@templaterTypesFor(id => Service, filename => uint8_slice_service.go, Name => UInt8Slice, Type => []uint8)
@templaterTypesFor(id => Service, filename => uint32_slice_service.go, Name => UInt32Slice, Type => []uint32)
@templaterTypesFor(id => Service, filename => uint64_slice_service.go, Name => UInt64Slice, Type => []uint64)
*/
//
package dime
