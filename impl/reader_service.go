package impl

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"github.com/influx6/dime/services"
)

//==========================================================================================================

// ReadStdoutService returns a new instance of a LineReaderService that uses os.Stdin as reader.
func ReadStdoutService(buffer int, maxWaitingTime time.Duration) *LineReaderService {
	return NewLineReaderService(buffer, maxWaitingTime, os.Stdout)
}

// ReadStdinService returns a new instance of a LineReaderService that uses os.Stdin as reader.
func ReadStdinService(buffer int, maxWaitingTime time.Duration) *LineReaderService {
	return NewLineReaderService(buffer, maxWaitingTime, os.Stdin)
}

// ReadStderrService returns a new instance of a LineReaderService that uses os.Stderr as service.
func ReadStderrService(buffer int, maxWaitingTime time.Duration) *LineReaderService {
	return NewLineReaderService(buffer, maxWaitingTime, os.Stderr)
}

//==========================================================================================================

// ReaderService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type ReaderService struct {
	pub         *services.BytesDistributor
	pubErrs     *services.ErrorDistributor
	reader      io.Reader
	maxBuffSize int
	stopped     chan struct{}
	incoming    chan []byte
}

// NewReaderService returns a new instance of a StdOutService.
func NewReaderService(maxReadBufferSize int, maxWaitingTime time.Duration, reader io.Reader) *ReaderService {
	pub := services.NewBytesDistributor(0, maxWaitingTime)
	pubErr := services.NewErrorDistributor(0, maxWaitingTime)

	stdServ := ReaderService{
		pub:         pub,
		pubErrs:     pubErr,
		reader:      reader,
		maxBuffSize: maxReadBufferSize,
		stopped:     make(chan struct{}, 0),
		incoming:    make(chan []byte, 0),
	}

	go stdServ.runReader()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *ReaderService) Done() <-chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *ReaderService) Stop() error {
	close(std.stopped)

	// Stop subscription delivery.
	std.pub.Stop()
	std.pubErrs.Stop()

	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *ReaderService) Write(in <-chan []byte) error {
	return ErrNotSupported
}

// ReadErrors returns a channel for reading error information to a listener.
func (std *ReaderService) ReadErrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *ReaderService) Read() (<-chan []byte, error) {
	mc := make(chan []byte, 0)

	std.pub.Subscribe(mc)

	return mc, nil
}

// runReader reads continously from the reader provider till io.EOF.
func (std *ReaderService) runReader() {
	reader := bufio.NewReader(std.reader)

	t := time.NewTimer(readerCloseCheckDuration)
	defer t.Stop()

	{
		for {
			// Validate that service has not be closed:
			select {
			case <-std.stopped:
				return
			case <-t.C:
				t.Reset(readerCloseCheckDuration)
				break
			}

			buf := make([]byte, std.maxBuffSize)

			_, err := reader.Read(buf)
			if err != nil {
				std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)

				if err == io.EOF {
					if len(buf) != 0 {
						std.pub.Publish(buf)
					}

					go std.Stop()
					return
				}

				continue
			}

			std.pub.Publish(buf)
			buf = nil
		}
	}
}

//==========================================================================================================

// LineReaderService implements a dime.MonoByteService for reading from os.Stdin and writing to os.Stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type LineReaderService struct {
	pub      *services.BytesDistributor
	pubErrs  *services.ErrorDistributor
	reader   io.Reader
	stopped  chan struct{}
	incoming chan []byte
}

// NewLineReaderService returns a new instance of a StdOutService.
func NewLineReaderService(buffer int, maxWaitingTime time.Duration, reader io.Reader) *LineReaderService {
	pub := services.NewBytesDistributor(buffer, maxWaitingTime)
	pubErr := services.NewErrorDistributor(buffer, maxWaitingTime)

	stdServ := LineReaderService{
		pub:      pub,
		pubErrs:  pubErr,
		reader:   reader,
		stopped:  make(chan struct{}, 0),
		incoming: make(chan []byte, 0),
	}

	go stdServ.runLineReader()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *LineReaderService) Done() <-chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *LineReaderService) Stop() error {
	close(std.stopped)

	std.pub.CloseAllSubs()
	std.pubErrs.CloseAllSubs()

	// Clear all pending subscribers.
	std.pub.Clear()
	std.pubErrs.Clear()

	// Stop subscription delivery.
	std.pub.Stop()
	std.pubErrs.Stop()

	return nil
}

// Write accepts a channel which data will be read from delivery data into the writer.
func (std *LineReaderService) Write(in <-chan []byte) error {
	return ErrNotSupported
}

// LineReaderrors returns a channel for reading error information to a listener.
func (std *LineReaderService) LineReaderrors() <-chan error {
	mc := make(chan error, 10)

	std.pubErrs.Subscribe(mc)

	return mc
}

// Read returns a channel for sending information to a listener.
func (std *LineReaderService) Read() (<-chan []byte, error) {
	mc := make(chan []byte, 0)

	std.pub.Subscribe(mc)

	return mc, nil
}

// runLineReader reads continously from the LineReader provider till io.EOF.
func (std *LineReaderService) runLineReader() {
	lineReader := bufio.NewReader(std.reader)

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
				break
			}

			// Have we reached the end of a LineReader?
			data, isPrefix, err := lineReader.ReadLine()
			if err != nil {
				std.pubErrs.PublishDeadline(err, errorWriteAcceptTimeout)

				if err == io.EOF {
					if len(data) != 0 && buf.Len() != 0 {
						buf.Write(data)
						std.pub.Publish(buf.Bytes())
						buf.Reset()

						go std.Stop()
						return
					}

					if len(data) != 0 {
						std.pub.Publish(data)
					}

					go std.Stop()
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
