package impl

import (
	"io"
	"os"
	"time"

	"github.com/influx6/dime/services"
)

// WriteStdoutService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type WriteStdoutService struct {
	pubErrs *services.ErrorDistributor

	stopped  chan struct{}
	incoming chan []byte
	writers  chan chan []byte
}

// NewWriteStdoutService returns a new instance of a StdOutService.
func NewWriteStdoutService(buffer int, maxWaitingTime time.Duration) *WriteStdoutService {
	pubErr := services.NewErrorDisributor(buffer, maxWaitingTime)

	defer pubErr.Start()

	stdServ := WriteStdoutService{
		pubErrs:  pubErr,
		stopped:  make(chan struct{}, 0),
		incoming: make(chan []byte, 0),
		writers:  make(chan chan []byte, 0),
	}

	go stdServ.runWriter()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *WriteStdoutService) Done() chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *WriteStdoutService) Stop() error {
	close(std.stopped)

	// Clear all pending subscribers.
	std.pubErrs.Clear()

	// Stop subscription delivery.
	std.pubErrs.Stop()

	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *WriteStdoutService) Write(in <-chan []byte) error {
	go std.lunchPublisher(in)
	return nil
}

// launches a go-routine to write data into publisher.
func (std *WriteStdoutService) lunchPublisher(in <-chan []byte) {
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

// ReadErrors returns a channel for reading error information to a listener.
func (std *WriteStdoutService) ReadErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *WriteStdoutService) Read() (<-chan []byte, error) {
	return nil, ErrNotSupported
}

// runWriter handles the internal processing of  writing data into provided writer.
func (std *WriteStdoutService) runWriter() {
	{
		for {
			select {
			case <-std.stopped:
				return
			case data, ok := <-std.incoming:
				if !ok {
					return
				}

				written, err := os.Stdout.Write(data)
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
