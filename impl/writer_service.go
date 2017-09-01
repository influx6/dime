package impl

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/influx6/dime/services"
)

//==========================================================================================================

// WriteMultiStdoutService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteMultiStdoutService(buffer int, maxWaitingTime time.Duration) *MultiWriterService {
	return NewMultiWriterService(buffer, maxWaitingTime, os.Stdout)
}

// WriteMultiStdinService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteMultiStdinService(buffer int, maxWaitingTime time.Duration) *MultiWriterService {
	return NewMultiWriterService(buffer, maxWaitingTime, os.Stdin)
}

// WriteMultiStderrService returns a new instance of a WriteService that uses os.Stderr as service.
func WriteMultiStderrService(buffer int, maxWaitingTime time.Duration) *MultiWriterService {
	return NewMultiWriterService(buffer, maxWaitingTime, os.Stderr)
}

// WriteSingleStdoutService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteSingleStdoutService(buffer int, maxWaitingTime time.Duration) *SingleWriterService {
	return NewSingleWriterService(buffer, maxWaitingTime, os.Stdout)
}

// WriteSingleStdinService returns a new instance of a WriteService that uses os.Stdin as reader.
func WriteSingleStdinService(buffer int, maxWaitingTime time.Duration) *SingleWriterService {
	return NewSingleWriterService(buffer, maxWaitingTime, os.Stdin)
}

// WriteSingleStderrService returns a new instance of a WriteService that uses os.Stderr as service.
func WriteSingleStderrService(buffer int, maxWaitingTime time.Duration) *SingleWriterService {
	return NewSingleWriterService(buffer, maxWaitingTime, os.Stderr)
}

//==========================================================================================================

// SingleWriterService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type SingleWriterService struct {
	pubErrs *services.ErrorDistributor

	writer   io.Writer
	stopped  chan struct{}
	done     chan struct{}
	incoming chan []byte
	writers  chan chan []byte
	locked   bool
	wg       sync.WaitGroup
	rw       sync.Mutex
	closed   bool
}

// NewSingleWriterService returns a new instance of a StdOutService.
func NewSingleWriterService(buffer int, maxWaitingTime time.Duration, w io.Writer) *SingleWriterService {
	pubErr := services.NewErrorDistributor(buffer)

	stdServ := SingleWriterService{
		writer:   w,
		pubErrs:  pubErr,
		stopped:  make(chan struct{}, 0),
		done:     make(chan struct{}, 0),
		incoming: make(chan []byte, 0),
		writers:  make(chan chan []byte, 0),
	}

	stdServ.wg.Add(2)
	go stdServ.runWriter()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *SingleWriterService) Done() <-chan struct{} {
	return std.done
}

// Stop ends all operations of the service.
func (std *SingleWriterService) Stop() error {
	std.rw.Lock()

	if std.closed {
		std.rw.Unlock()
		return nil
	}
	std.rw.Unlock()

	close(std.stopped)

	std.wg.Wait()

	// Stop subscription delivery.
	std.pubErrs.Stop()
	std.closed = true

	close(std.done)
	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *SingleWriterService) Write(in <-chan []byte) error {
	if std.locked {
		return ErrOnlySingleWriteChannelSupported
	}

	go std.lunchPublisher(in)
	std.locked = true

	return nil
}

// ReadErrors returns a channel for reading error information to a listener.
func (std *SingleWriterService) ReadErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *SingleWriterService) Read() (<-chan []byte, error) {
	return nil, ErrNotSupported
}

// launches a go-routine to write data into publisher.
func (std *SingleWriterService) lunchPublisher(in <-chan []byte) {
	defer std.wg.Done()

	t := time.NewTimer(writerWaitDuration)
	defer t.Stop()

	for {
		select {
		case <-std.stopped:
			return
		case data, ok := <-in:
			if !ok {
				go std.Stop()
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
func (std *SingleWriterService) WriteErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// runWriter handles the internal processing of  writing data into provided writer.
func (std *SingleWriterService) runWriter() {
	defer std.wg.Done()

	{
		for {
			select {
			case <-std.stopped:
				return
			case data := <-std.incoming:
				written, err := std.writer.Write(data)
				if err != nil {
					std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)

					go std.Stop()
					return
				}

				if written != len(data) {
					std.pubErrs.PublishDeadline(io.ErrShortWrite, errorWriteAcceptTimeout)
				}
			}
		}
	}
}

//==========================================================================================================

// MultiWriterService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type MultiWriterService struct {
	pubErrs *services.ErrorDistributor

	writer   io.Writer
	stopped  chan struct{}
	done     chan struct{}
	incoming chan []byte
	writers  chan chan []byte
	wg       sync.WaitGroup
	rw       sync.Mutex
	closed   bool
}

// NewMultiWriterService returns a new instance of a StdOutService.
func NewMultiWriterService(buffer int, maxWaitingTime time.Duration, w io.Writer) *MultiWriterService {
	pubErr := services.NewErrorDistributor(buffer)

	stdServ := MultiWriterService{
		writer:   w,
		pubErrs:  pubErr,
		stopped:  make(chan struct{}, 0),
		done:     make(chan struct{}, 0),
		incoming: make(chan []byte, 0),
		writers:  make(chan chan []byte, 0),
	}

	stdServ.wg.Add(1)
	go stdServ.runWriter()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *MultiWriterService) Done() <-chan struct{} {
	return std.done
}

// Stop ends all operations of the service.
func (std *MultiWriterService) Stop() error {
	std.rw.Lock()

	if std.closed {
		std.rw.Unlock()
		return nil
	}
	std.rw.Unlock()

	close(std.stopped)

	std.wg.Wait()

	// Stop subscription delivery.
	std.pubErrs.Stop()
	std.closed = true

	close(std.done)

	return nil
}

// ReadErrors returns a channel for reading error information to a listener.
func (std *MultiWriterService) ReadErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *MultiWriterService) Read() (<-chan []byte, error) {
	return nil, ErrNotSupported
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *MultiWriterService) Write(in <-chan []byte) error {
	std.wg.Add(1)
	go std.lunchPublisher(in)
	return nil
}

// launches a go-routine to write data into publisher.
func (std *MultiWriterService) lunchPublisher(in <-chan []byte) {
	defer std.wg.Done()

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
func (std *MultiWriterService) WriteErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// runWriter handles the internal processing of  writing data into provided writer.
func (std *MultiWriterService) runWriter() {
	defer std.wg.Done()

	{
		for {
			select {
			case <-std.stopped:
				return
			case data, ok := <-std.incoming:
				if !ok {
					go std.Stop()
					return
				}

				written, err := std.writer.Write(data)
				if err != nil {
					std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)

					go std.Stop()
					return
				}

				if written != len(data) {
					std.pubErrs.PublishDeadline(io.ErrShortWrite, errorWriteAcceptTimeout)
				}
			}
		}
	}
}
