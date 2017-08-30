package functions

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/influx6/dime/impl"
	"github.com/influx6/dime/json"
	"github.com/influx6/faux/metrics"

	"github.com/influx6/dime/services"
)

var (
	defaultWaitForSending = 3 * time.Second
)

// StdFunction returns a ByteFunctor instance which processing data coming from stdin
// and receiving both data and error into stdout and stderr appropriately.
func StdFunction(maxTimeToAbortReply time.Duration, bx ByteFunction) (*ByteFunctor, error) {
	reader := impl.NewReadStdInService(0, maxTimeToAbortReply)
	writer := impl.NewWriteStdoutService(0, maxTimeToAbortReply)
	errs := impl.NewWriteStderrService(0, maxTimeToAbortReply)

	return NewByteFunctor(
		UseByteFunctorFunction(bx),
		SetByteFunctorServiceClosing(true),
		UseByteFunctorStreams(reader, writer, errs),
	)
}

// RWFunction returns a new instance of a ByteFunctor using the associated reader and writers.
func RWFunction(reader io.Reader, outw, errw io.Writer, bx ByteFunction) (*ByteFunctor, error) {
	rw := impl.NewReaderService(0, defaultWaitForSending, reader)
	ww := impl.NewWriterService(0, defaultWaitForSending, outw)
	we := impl.NewWriterService(0, defaultWaitForSending, errw)

	return NewByteFunctor(
		UseByteFunctorFunction(bx),
		SetByteFunctorServiceClosing(true),
		UseByteFunctorStreams(rw, ww, we),
	)
}

//=============================================================================================

// errors
var (
	ErrNoFunctionSet = errors.New("Processing Function required")
	ErrNoInputSet    = errors.New("Input MonoBytesService required")
	ErrNoOutputSet   = errors.New("Output MonoBytesService required")
	ErrNoErrputSet   = errors.New("Error MonoBytesService required")
)

// ByteFunctorOption defines a function type which recieves a *ByteFunctor for field setting.
type ByteFunctorOption func(*ByteFunctor)

// ByteFunction defines a function type for the ByteFunctor.
type ByteFunction func(services.CancelContext, <-chan []byte, chan<- []byte, chan<- error)

// SetByteFunctorServiceClosing sets the ByteFunctor to close all attached services when it's Close() method is called.
func SetByteFunctorServiceClosing(flag bool) ByteFunctorOption {
	return func(bm *ByteFunctor) {
		bm.enableServiceClosing = flag
	}
}

// MetricByteFunctorFunction returns a ByteFunctorOption that wraps the provided ByteFunction
// to collect metrics on total run time, stack traces and others based on provided id.
func MetricByteFunctorFunction(id string, m metrics.Metrics, fx ByteFunction) ByteFunctorOption {
	return func(bm *ByteFunctor) {
		var executionCount int

		bm.fx = func(ctx services.CancelContext, in <-chan []byte, out chan<- []byte, errs chan<- error) {
			total := executionCount
			executionCount++

			defer m.Emit(metrics.WithFields(metrics.Fields{
				"id":            id,
				"totalExecuted": total,
				"executionId":   executionCount,
				"area":          "functors.stack",
				"track":         "functor_executions",
			}).Trace("Executing ByteFunctor").End())

			inView := services.BytesView(ctx, 100*time.Millisecond, func(received []byte) {
				m.Emit(metrics.WithFields(metrics.Fields{
					"id":          id,
					"executionId": executionCount,
					"area":        "functors.incoming",
					"track":       "functor_executions",
					"data":        received,
				}))
			}, in)

			outView := services.BytesSinkView(ctx, false, 100*time.Millisecond, func(sent []byte) {
				m.Emit(metrics.WithFields(metrics.Fields{
					"id":          id,
					"executionId": executionCount,
					"area":        "functors.outgoing",
					"track":       "functor_executions",
					"data":        sent,
				}))
			}, out)

			errView := services.ErrorSinkView(ctx, false, 100*time.Millisecond, func(whyErr error) {
				m.Emit(metrics.WithFields(metrics.Fields{
					"id":          id,
					"executionId": executionCount,
					"area":        "functors.outgoing.error",
					"track":       "functor_executions",
					"data":        whyErr,
				}))
			}, errs)

			defer close(outView)
			defer close(errView)

			fx(ctx, inView, outView, errView)
		}
	}
}

// UseByteFunctorFunction sets the processing function for the ByteFunctor.
func UseByteFunctorFunction(fx ByteFunction) ByteFunctorOption {
	return func(bm *ByteFunctor) {
		bm.fx = fx
	}
}

// UseByteFunctorStreams  sets the reader and writers for a giving ByteFunctor.
func UseByteFunctorStreams(inreader, outwriter, errwriter services.MonoBytesService) ByteFunctorOption {
	return func(bm *ByteFunctor) {
		bm.stdin = inreader
		bm.stdout = outwriter
		bm.stderr = errwriter
	}
}

// ByteFunctor defines a struct which hosts a giving function, sending data read from stdin
// into the function for processing and retriving data continously through a output channel.
type ByteFunctor struct {
	fx                   ByteFunction
	enableServiceClosing bool
	stdin                services.MonoBytesService
	stdout               services.MonoBytesService
	stderr               services.MonoBytesService
	closer               chan struct{}
	wg                   sync.WaitGroup
}

// NewByteFunctor returns a new instance of a ByteFunctor.
func NewByteFunctor(option ...ByteFunctorOption) (*ByteFunctor, error) {
	bm := &ByteFunctor{}

	for _, op := range option {
		op(bm)
	}

	if bm.fx == nil {
		return nil, ErrNoFunctionSet
	}

	if bm.stderr == nil {
		return nil, ErrNoErrputSet
	}

	if bm.stdout == nil {
		return nil, ErrNoOutputSet
	}

	if bm.stdin == nil {
		return nil, ErrNoInputSet
	}

	bm.wg.Add(1)
	go bm.run()

	return bm, nil
}

// Done returns a channel for receiving done/cancel notifications.
func (bx *ByteFunctor) Done() <-chan struct{} {
	return bx.closer
}

// Close cancels any current running operation.
func (bx *ByteFunctor) Close() {
	close(bx.closer)

	bx.wg.Wait()

	if !bx.enableServiceClosing {
		return
	}

	bx.stdin.Stop()
	bx.stdout.Stop()
	bx.stderr.Stop()
}

func (bx *ByteFunctor) run() {
	defer bx.wg.Done()

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

	if err := bx.stderr.Write(json.ErrorBecomeBytesWithJSON(bx, outErr)); err != nil {
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
