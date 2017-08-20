package impl

import (
	"time"

	"github.com/influx6/dime/services"
)

// StdOutService implements a dime.MonoByteService for reading an writing to stdout.
// It allows us expose the stdout as a stream of continouse bytes to be read and written to.
type StdOutService struct {
	pub     *services.BytesDistributor
	pubErrs *services.ErrorDistributor

	errs     chan error
	stopped  chan struct{}
	incoming chan []byte
}

// NewStdoutService returns a new instance of a StdOutService.
func NewStdoutService(buffer int, maxWaitingTime time.Duration) *StdOutService {
	pub := services.NewBytesDisributor(buffer, maxWaitingTime)
	pubErr := services.NewErrorDisributor(buffer, maxWaitingTime)

	defer pubErr.Start()
	defer pub.Start()

	stdServ := StdOutService{
		pub:      pub,
		pubErrs:  pubErr,
		incoming: make(chan []byte, 0),
		stopped:  make(chan struct{}, 0),
	}

	go stdServ.run()

	return &stdServ
}

// Done returns a channel which will be closed once the service is stopped.
func (std *StdOutService) Done() chan struct{} {
	return std.stopped
}

// Stop ends all operations of the service.
func (std *StdOutService) Stop() error {

	// Clear all pending subscribers.
	std.pub.Clear()
	std.pubErrs.Clear()

	// Stop subscription delivery.
	std.pub.Stop()
	std.pubErrs.Stop()

	close(std.stopped)

	return nil
}

// Write accepts a channel which data will be read from till data is closed.
func (std *StdOutService) Write(in <-chan []byte) error {

	return nil
}

// ReadErrors returns a channel for sending information into stdout.
func (std *StdOutService) ReadErrors() (<-chan error, error) {
	mc := make(chan error, 0)

	std.pubErrs.Subscribe(mc)

	return mc, nil
}

// Read returns a channel for sending information into stdout.
func (std *StdOutService) Read() (<-chan []byte, error) {
	mc := make(chan []byte, 0)

	std.pub.Subscribe(mc)

	return mc, nil
}

// run handles the internal processing of reading and writing sync into
// stdout.
func (std *StdOutService) run() {
  {
    for {
      select {
      case <-std.stopped:
        return
      case 
      }
    }
  }
}
