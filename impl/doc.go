// Package impl provides implmenetation of structures that implement specific dime interface types.
//
// @templater(id => Common, gen => Partial.Go, file => _common.tml)
// @templater(id => ImplService, gen => Partial.Go, file => _service_impl.tml)
//
// @templaterTypesFor(id => Common, filename => common.go)
// @templaterTypesFor(id => ImplService, filename => stdin_reader_service.go, Name => ReadStdInService, Type => []byte)
// @templaterTypesFor(id => ImplService, filename => stdin_writer_service.go, Name => WriteStdInService, Type => []byte)
// @templaterTypesFor(id => ImplService, filename => stdout_reader_service.go, Name => ReadStdoutService, Type => []byte)
// @templaterTypesFor(id => ImplService, filename => stdout_writer_service.go, Name => WriteStdoutService, Type => []byte)
// @templaterTypesFor(id => ImplService, filename => stderr_reader_service.go, Name => ReadStderrService, Type => []byte)
// @templaterTypesFor(id => ImplService, filename => stderr_writer_service.go, Name => WriteStderrService, Type => []byte)
//
package impl
