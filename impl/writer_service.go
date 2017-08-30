package impl

import (
	"io"
	"os"
	"time"

	"github.com/influx6/dime/services"
)

//==========================================================================================================

// WriteStdoutService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteStdoutService(buffer int, maxWaitingTime time.Duration) *WriterService {
	return NewWriterService(buffer, maxWaitingTime, os.Stdout)
}

// WriteStdinService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteStdinService(buffer int, maxWaitingTime time.Duration) *WriterService {
	return NewWriterService(buffer, maxWaitingTime, os.Stdin)
}

// WriteStderrService returns a new instance of a WriteService that uses os.Stderr as service.
func WriteStderrService(buffer int, maxWaitingTime time.Duration) *WriterService {
	return NewWriterService(buffer, maxWaitingTime, os.Stderr)
}

//==========================================================================================================

// WriterService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type WriterService struct {
	pubErrs *services.ErrorDistributor

	writer   io.Writer
	stopped  chan struct{}
	incoming chan []byte
	writers  chan chan []byte
}

// NewWriterService returns a new instance of a StdOutService.
func NewWriterService(buffer int, maxWaitingTime time.Duration, w io.Writer) *WriterService {
	pubErr := services.NewErrorDistributor(buffer, maxWaitingTime)

	defer pubErr.Start()

	stdServ := WriterService{
		writer:   w,
		pubErrs:  pubErr,
		stopped:  make(chan struct{}, 0),
		incoming: make(chan []byte, 0),
		writers:  make(chan chan []byte, 0),
	}

	go stdServ.runWriter()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *WriterService) Done() <-chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *WriterService) Stop() error {
	close(std.stopped)

	// Clear all pending subscribers.
	std.pubErrs.CloseAllSubs()
	std.pubErrs.Clear()

	// Stop subscription delivery.
	std.pubErrs.Stop()

	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *WriterService) Write(in <-chan []byte) error {
	go std.lunchPublisher(in)
	return nil
}

// launches a go-routine to write data into publisher.
func (std *WriterService) lunchPublisher(in <-chan []byte) {
	t := time.NewTimer(writerWaitDuration)
	defer t.Stop()

	for {
		select {
		case <-std.stopped:
			return
		case data, ok := <-in:
			if !ok {
				return
			}

			std.incoming <- data
		case <-t.C:
			t.Reset(writerWaitDuration)
			continue
		}
	}
}

// WriteErrors returns a channel for reading error information to a listener.
func (std *WriterService) WriteErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// runWriter handles the internal processing of  writing data into provided writer.
func (std *WriterService) runWriter() {
	{
		for {
			select {
			case <-std.stopped:
				return
			case data, ok := <-std.incoming:
				if !ok {
					std.pubErrs.CloseAllSubs()
					return
				}

				written, err := std.writer.Write(data)
				if err != nil {
					std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)
					continue
				}

				if written != len(data) {
					std.pubErrs.PublishDeadline(io.ErrShortWrite, errorWriteAcceptTimeout)
				}
			}
		}
	}
}
