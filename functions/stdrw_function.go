package functions

import (
	"time"

	"github.com/influx6/dime/impl"

	"github.com/influx6/dime/services"
)

// StdFunction returns a ByteRWFunction instance which processing data coming from stdin
// and receiving both data and error into stdout and stderr appropriately.
func StdFunction(maxTimeToAbortReply time.Duration, bx ByteFunction) *ByteRWFunction {
	reader := impl.NewReadStdInService(0, maxTimeToAbortReply)
	writer := impl.NewWriteStdoutService(0, maxTimeToAbortReply)
	errs := impl.NewWriteStderrService(0, maxTimeToAbortReply)

	return NewByteRWFunction(reader, writer, errs, bx)
}

//=============================================================================================

// ByteFunction defines a function type for the ByteRWFunction.
type ByteFunction func(services.CancelContext, <-chan []byte, chan<- []byte, chan<- error)

// ByteRWFunction defines a struct which hosts a giving function, sending data read from stdin
// into the function for processing and retriving data continously through a output channel.
type ByteRWFunction struct {
	fx     ByteFunction
	stdin  services.MonoBytesService
	stdout services.MonoBytesService
	stderr services.MonoBytesService
	closer chan struct{}
}

// NewByteRWFunction returns a new instance of a ByteRWFunction.
func NewByteRWFunction(insrv, outsrv services.MonoBytesService, errsrv services.MonoBytesService, fnx ByteFunction) *ByteRWFunction {
	bm := ByteRWFunction{
		stdin:  insrv,
		stdout: outsrv,
		stderr: errsrv,
		fx:     fnx,
	}

	go bm.run()

	return &bm
}

// Done returns a channel for receiving done/cancel notifications.
func (bx *ByteRWFunction) Done() <-chan struct{} {
	return bx.closer
}

// Cancel cancels any current running operation.
func (bx *ByteRWFunction) Cancel() {
	close(bx.closer)
}

func (bx *ByteRWFunction) run() {
	t := time.NewTimer(20 * time.Millisecond)
	defer t.Stop()

	incoming, err := bx.stdin.Read()
	if err != nil {
		panic(err)
	}

	output := make(chan []byte)
	outErr := make(chan error)

	defer close(output)
	defer close(outErr)

	if err := bx.stdout.Write(output); err != nil {
		panic(err)
	}

	if err := bx.stderr.Write(outErr); err != nil {
		panic(err)
	}

	for {
		bx.fx(bx, incoming, output, outErr)

		select {
		case <-bx.closer:
			return
		case <-t.C:
			t.Reset(20 * time.Millisecond)
			continue
		}
	}
}
