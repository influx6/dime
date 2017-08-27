// Package json implements a adapter functions for golang base types and their slice versions.
//
//
/*

  @templater(id => Service, gen => Partial.Go, file => _types.tml)

Below are types annotations used to generate different Service types based on a giving type.

  @templaterTypesFor(id => Service, filename => map_json_adapter.go, Name => Map, Type => map[string]string)
  @templaterTypesFor(id => Service, filename => map_slice_json_adapter.go, Name => MapSlice, Type => []map[string]string)

  @templaterTypesFor(id => Service, filename => map_of_any_json_adapter.go, Name => MapOfAny, Type => map[interface{}]interface{})
  @templaterTypesFor(id => Service, filename => map_of_any_slice_json_adapter.go, Name => MapOfAnySlice, Type => []map[interface{}]interface{})

  @templaterTypesFor(id => Service, filename => interface_json_adapter.go, Name => Interface, Type => interface{})
  @templaterTypesFor(id => Service, filename => interface_slice_json_adapter.go, Name => InterfaceSlice, Type => []interface{})

  @templaterTypesFor(id => Service, filename => string_json_adapter.go, Name => String, Type => string)
  @templaterTypesFor(id => Service, filename => string_slice_json_adapter.go, Name => StringSlice, Type => []string)

  @templaterTypesFor(id => Service, filename => error_json_adapter.go, Name => Error, Type => error)
  @templaterTypesFor(id => Service, filename => error_slice_json_adapter.go, Name => ErrorSlice, Type => []error)

  @templaterTypesFor(id => Service, filename => bool_json_adapter.go, Name => Bool, Type => bool)
  @templaterTypesFor(id => Service, filename => bool_slice_json_adapter.go, Name => BoolSlice, Type => []bool)

  @templaterTypesFor(id => Service, NoAdapter => true, filename => byte_json_adapter.go, Name => Byte, Type => byte)
  @templaterTypesFor(id => Service, NoAdapter => true, filename => rune_json_adapter.go, Name => Rune, Type => rune)
  @templaterTypesFor(id => Service, NoAdapter => true, filename => bytes_json_adapter.go, Name => Bytes, Type => []byte)

  @templaterTypesFor(id => Service, filename => float64_json_adapter.go, Name => Float64, Type => float64)
  @templaterTypesFor(id => Service, filename => float64_slice_json_adapter.go, Name => Float64Slice, Type => []float64)

  @templaterTypesFor(id => Service, filename => float32_json_adapter.go, Name => Float32, Type => float32)
  @templaterTypesFor(id => Service, filename => float32_slice_json_adapter.go, Name => Float32Slice, Type => []float32)

  @templaterTypesFor(id => Service, filename => complex64_json_adapter.go, Name => Complex64, Type => complex64)
  @templaterTypesFor(id => Service, filename => complex128_json_adapter.go, Name => Complex128, Type => complex128)

  @templaterTypesFor(id => Service, filename => complex64_slice_json_adapter.go, Name => Complex64Slice, Type => []complex64)
  @templaterTypesFor(id => Service, filename => complex128_slice_json_adapter.go, Name => Complex128Slice, Type => []complex128)

  @templaterTypesFor(id => Service, filename => int_json_adapter.go, Name => Int, Type => int)
  @templaterTypesFor(id => Service, filename => int8_json_adapter.go, Name => Int8, Type => int8)
  @templaterTypesFor(id => Service, filename => int16_json_adapter.go, Name => Int16, Type => int16)
  @templaterTypesFor(id => Service, filename => int32_json_adapter.go, Name => Int32, Type => int32)
  @templaterTypesFor(id => Service, filename => int64_json_adapter.go, Name => Int64, Type => int64)

  @templaterTypesFor(id => Service, filename => int_slice_json_adapter.go, Name => IntSlice, Type => []int)
  @templaterTypesFor(id => Service, filename => int8_slice_json_adapter.go, Name => Int8Slice, Type => []int8)
  @templaterTypesFor(id => Service, filename => int16_slice_json_adapter.go, Name => Int16Slice, Type => []int16)
  @templaterTypesFor(id => Service, filename => int32_slice_json_adapter.go, Name => Int32Slice, Type => []int32)
  @templaterTypesFor(id => Service, filename => int64_slice_json_adapter.go, Name => Int64Slice, Type => []int64)

  @templaterTypesFor(id => Service, filename => uint_json_adapter.go, Name => UInt, Type => uint)
  @templaterTypesFor(id => Service, filename => uint8_json_adapter.go, Name => UInt8, Type => uint8)
  @templaterTypesFor(id => Service, filename => uint32_json_adapter.go, Name => UInt32, Type => uint32)
  @templaterTypesFor(id => Service, filename => uint64_json_adapter.go, Name => UInt64, Type => uint64)

  @templaterTypesFor(id => Service, filename => uint_slice_json_adapter.go, Name => UIntSlice, Type => []uint)
  @templaterTypesFor(id => Service, filename => uint8_slice_json_adapter.go, Name => UInt8Slice, Type => []uint8)
  @templaterTypesFor(id => Service, filename => uint32_slice_json_adapter.go, Name => UInt32Slice, Type => []uint32)
  @templaterTypesFor(id => Service, filename => uint64_slice_json_adapter.go, Name => UInt64Slice, Type => []uint64)

*/
//
package json
