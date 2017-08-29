package impl

import (
	"bufio"
	"bytes"
	"io"
	"time"

	"github.com/influx6/dime/services"
)

var (
	peekLen = 10
)

// StdService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type StdService struct {
	pub     *services.BytesDistributor
	pubErrs *services.ErrorDistributor

	readFrom io.Reader
	writeTo  io.Writer
	stopped  chan struct{}
	writers  chan chan []byte
	incoming chan []byte
}

// NewStdService returns a new instance of a StdOutService.
func NewStdService(buffer int, maxWaitingTime time.Duration, readFrom io.Reader, writeTo io.Writer) *StdService {
	pub := services.NewBytesDistributor(buffer, maxWaitingTime)
	pubErr := services.NewErrorDistributor(buffer, maxWaitingTime)

	defer pubErr.Start()
	defer pub.Start()

	stdServ := StdService{
		pub:      pub,
		pubErrs:  pubErr,
		writeTo:  writeTo,
		readFrom: readFrom,
		incoming: make(chan []byte, 0),
		stopped:  make(chan struct{}, 0),
		writers:  make(chan chan []byte, 0),
	}

	go stdServ.runReader()
	go stdServ.runWriter()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *StdService) Done() chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *StdService) Stop() error {
	close(std.stopped)

	// Clear all pending subscribers.
	std.pub.Clear()
	std.pubErrs.Clear()

	// Stop subscription delivery.
	std.pub.Stop()
	std.pubErrs.Stop()

	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *StdService) Write(in <-chan []byte) error {
	go std.lunchPublisher(in)
	return nil
}

// launches a go-routine to write data into publisher.
func (std *StdService) lunchPublisher(in <-chan []byte) {
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
func (std *StdService) ReadErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *StdService) Read() (<-chan []byte, error) {
	mc := make(chan []byte, 0)

	std.pub.Subscribe(mc)

	return mc, nil
}

// runReader reads continously from the reader provider till io.EOF.
func (std *StdService) runReader() {
	if std.readFrom == nil {
		return
	}

	reader := bufio.NewReader(std.readFrom)

	t := time.NewTimer(readerCloseCheckDuration)
	defer t.Stop()

	var buf bytes.Buffer
	var multiline bool

	{
		for {

			// Validate that service has not be closed:
			select {
			case <-std.stopped:
				return
			case <-t.C:
				t.Reset(readerCloseCheckDuration)
				// break
			}

			// Have we reached the end of a reader?
			data, isPrefix, err := reader.ReadLine()
			if err != nil {
				std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)

				if err == io.EOF {
					if len(data) != 0 && buf.Len() != 0 {
						buf.Write(data)
						std.pub.Publish(buf.Bytes())
						buf.Reset()
						return
					}

					if len(data) != 0 {
						std.pub.Publish(data)
					}

					return
				}

				continue
			}

			if multiline && isPrefix {
				buf.Write(data)
				continue
			}

			if !multiline && isPrefix {
				multiline = true
				buf.Write(data)
				continue
			}

			if multiline && !isPrefix {
				multiline = false

				buf.Write(data)

				std.pub.Publish(buf.Bytes())
				buf.Reset()
				continue
			}

			std.pub.Publish(data)
		}
	}
}

// runWriter handles the internal processing of  writing data into provided writer.
func (std *StdService) runWriter() {
	if std.writeTo == nil {
		return
	}

	{
		for {
			select {
			case <-std.stopped:
				return
			case data, ok := <-std.incoming:
				if !ok {
					return
				}

				_, err := std.writeTo.Write(data)
				if err != nil {
					std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)
					continue
				}
			}
		}
	}
}
