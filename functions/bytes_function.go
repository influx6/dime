package functions

import (
	"errors"
	"io"
	"time"

	"github.com/influx6/dime/impl"
	"github.com/influx6/faux/metrics"

	"github.com/influx6/dime/services"
)

// errors
var (
	ErrNoFunctionSet = errors.New("Processing Function required")
	ErrNoInputSet    = errors.New("Input MonoBytesService required")
	ErrNoOutputSet   = errors.New("Output MonoBytesService required")
	ErrNoErrputSet   = errors.New("Error MonoBytesService required")
)

var (
	defaultWaitForSending = 3 * time.Second
)

// ByteFunction defines a function type for the ByteFunctor.
type ByteFunction func(services.CancelContext, <-chan []byte, chan<- []byte, <-chan error)

// LaunchStdFunction returns a ByteFunctor instance which processing data coming from stdin
// and receiving both data and error into stdout and stderr appropriately.
func LaunchStdFunction(ctx services.CancelContext, bufferSize int, maxTimeToAbortReply time.Duration, bx ByteFunction) error {
	reader := impl.ReadStdinService(bufferSize, maxTimeToAbortReply)
	writer := impl.WriteSingleStdoutService(bufferSize, maxTimeToAbortReply)

	if err := launchFunction(ctx, bx, reader, writer); err != nil {
		return err
	}

	return nil
}

// LaunchReaderWriterFunction returns a new instance of a ByteFunctor using the associated reader and writers.
func LaunchReaderWriterFunction(ctx services.CancelContext, bufferSize int, reader io.Reader, writer io.Writer, bx ByteFunction) error {
	rw := impl.NewReaderService(bufferSize, defaultWaitForSending, reader)
	ww := impl.NewSingleWriterService(bufferSize, defaultWaitForSending, writer)

	if err := launchFunction(ctx, bx, rw, ww); err != nil {
		return err
	}

	return nil
}

func launchFunction(ctx services.CancelContext, bx ByteFunction, requestReader, responseWriter services.MonoBytesService) error {
	incoming, err := requestReader.Read()
	if err != nil {
		return err
	}

	output := make(chan []byte)

	if err := responseWriter.Write(output); err != nil {
		return err
	}

	bx(ctx, incoming, output, requestReader.ReadErrors())

	<-requestReader.Done()
	<-responseWriter.Done()

	return nil
}

//=============================================================================================

// WrapWithMetric returns a ByteFunctorOption that wraps the provided ByteFunction
// to collect metrics on total run time, stack traces and others based on provided id.
func WrapWithMetric(id string, m metrics.Metrics, fx ByteFunction) ByteFunction {
	var executionCount int

	return func(ctx services.CancelContext, in <-chan []byte, out chan<- []byte, errs <-chan error) {
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

		outView := services.BytesSinkView(ctx, true, 100*time.Millisecond, func(sent []byte) {
			m.Emit(metrics.WithFields(metrics.Fields{
				"id":          id,
				"executionId": executionCount,
				"area":        "functors.outgoing",
				"track":       "functor_executions",
				"data":        sent,
			}))
		}, out)

		errView := services.ErrorView(ctx, 100*time.Millisecond, func(whyErr error) {
			m.Emit(metrics.WithFields(metrics.Fields{
				"id":          id,
				"executionId": executionCount,
				"area":        "functors.outgoing.error",
				"track":       "functor_executions",
				"data":        whyErr,
			}))
		}, errs)

		fx(ctx, inView, outView, errView)
	}
}
