// Autogenerated using the moz templater annotation.
//
//
package services

import (
	"context"
	"sync/atomic"
	"time"
)

//go:generate moz generate-file -fromFile ./float64_service.go -toDir ./impl/float64

// Float64FromByteAdapter defines a function that that will take a channel of bytes and return a channel of float64.
type Float64FromByteAdapterWithContext func(context.Context, chan []byte) chan float64

// Float64ToByteAdapter defines a function that that will take a channel of bytes and return a channel of float64.
type Float64ToByteAdapter func(context.Context, chan float64) chan []byte

// Float64PartialCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed or the context expires returning whatever was collected, and closing the returning channel.
// This function does not guarantee complete data.
func Float64PartialCollect(ctx context.Context, waitTime time.Duration, in chan float64) chan []float64 {
	res := make(chan []float64, 0)

	go func() {
		var buffer []float64

		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				res <- buffer
				close(res)
				return
			case data, ok := <-in:
				if !ok {
					res <- buffer
					close(res)
					return
				}

				buffer = append(buffer, data)
				continue
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// Float64Collect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed, nothing will be returned if the channel given is not closed  or the context expires.
// Once done, returning channel is closed.
// This function guarantees complete data.
func Float64Collect(ctx context.Context, waitTime time.Duration, in chan float64) chan []float64 {
	res := make(chan []float64, 0)

	go func() {
		var buffer []float64

		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(res)
				return
			case data, ok := <-in:
				if !ok {
					res <- buffer
					close(res)
					return
				}

				buffer = append(buffer, data)
				continue
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// Float64Mutate defines a function which returns a channel where the items of the incoming channel
// are mutated based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func Float64Mutate(ctx context.Context, waitTime time.Duration, mutateFn func(float64) float64, in chan float64) chan float64 {
	res := make(chan float64, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(res)
				return
			case data, ok := <-in:
				if !ok {
					close(res)
					return
				}

				res <- mutateFn(data)
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// Float64Filter defines a function which returns a channel where the items of the incoming channel
// are filtered based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func Float64Filter(ctx context.Context, waitTime time.Duration, filterFn func(float64) bool, in chan float64) chan float64 {
	res := make(chan float64, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(res)
				return
			case data, ok := <-in:
				if !ok {
					close(res)
					return
				}

				if !filterFn(data) {
					continue
				}

				res <- data
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// Float64CollectUntil defines a function which returns a channel where the items of the incoming channel
// are buffered until the data matches a given requirement provided through a function. If the function returns true
// then currently buffered data is returned and a new buffer is created. This is useful for batch collection based on
// specific criteria. If the channel is closed before the criteria is met, what data is left is sent down the returned channel,
// closing that channel. If the context expires then data gathered is returned and returning channel is closed.
// This function guarantees some data to be delivered.
func Float64CollectUntil(ctx context.Context, waitTime time.Duration, condition func([]float64) bool, in chan float64) chan []float64 {
	res := make(chan []float64, 0)

	go func() {
		var buffer []float64

		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(res)
				return
			case data, ok := <-in:
				if !ok {
					res <- buffer
					close(res)
					return
				}

				buffer = append(buffer, data)

				// If we do not match the given criteria, then continue buffering.
				if condition(buffer) {
					continue
				}

				// We do match criteria, send buffered data, and reset buffer
				res <- buffer
				buffer = nil
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// Float64MergeWithoutOrder merges the incoming data from the float64 into a single stream of float64,
// merge collects data from all provided channels in turns, each giving a specified time to deliver, else
// skipped until another turn. Once all data is collected from each sender, then the data set is merged into
// a single slice and delivered to the returned channel.
// MergeWithoutOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64MergeWithoutOrder(ctx context.Context, maxWaitTime time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		var index int

		total := len(senders)
		filled := make(map[int]float64, 0)

		for {
			// if the current index has being filled, shift forward and reattempt loop.
			if _, ok := filled[index]; ok {
				index++
				continue
			}

			if len(filled) == total {
				var content []float64

				for _, item := range filled {
					content = append(content, item)
				}

				res <- content

				index = 0
				filled = make(map[int]float64, 0)
			}

			timer := time.NewTimer(maxWaitTime)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					close(res)
					timer.Stop()
					return
				}

				filled[index] = data
				index++
			}

			timer.Stop()
		}
	}()

	return res
}

// Float64MergeInOrder merges the incoming data from the float64 into a single stream of float64,
// merge collects data from all provided channels in turns, each giving a specified time to deliver, else
// skipped until another turn. Once all data is collected from each sender, then the data set is merged into
// a single slice and delivered to the returned channel.
// MergeInOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64MergeInOrder(ctx context.Context, maxWaitTime time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		var index int

		total := len(senders)
		filled := make(map[int]float64, 0)

		for {
			// if the current index has being filled, shift forward and reattempt loop.
			if _, ok := filled[index]; ok {
				index++
				continue
			}

			if len(filled) == total {
				var content []float64

				for index := range senders {
					item := filled[index]
					content = append(content, item)
				}

				res <- content

				index = 0
				filled = make(map[int]float64, 0)
			}

			timer := time.NewTimer(maxWaitTime)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					close(res)
					timer.Stop()
					return
				}

				filled[index] = data
				index++
			}

			timer.Stop()
		}
	}()

	return res
}

// Float64CombineParitallyWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// Float64CombineParitallyWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64CombinePartiallyWithoutOrder(ctx context.Context, maxItemWait time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]float64, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		closed := make(map[int]bool, 0)

		var sendersClosed int

		for {
			// if the current index has being filled, shift forward and re-attempt loop.
			if filled[index] || closed[index] {
				index++
				continue
			}

			if len(content) == total {
				res <- content

				index = 0
				filled = make(map[int]bool, 0)
				content = make([]float64, len(senders))
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					timer.Stop()

					sendersClosed++
					closed[index] = true
					continue
				}

				content = append(content, data)
				filled[index] = true
				index++
			}

			timer.Stop()
		}

	}()

	return res
}

// Float64CombineWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// Float64CombineWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64CombineWithoutOrder(ctx context.Context, maxItemWait time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]float64, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] {
				index++
				continue
			}

			if len(content) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]float64, len(senders))
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					close(res)
					timer.Stop()
					return
				}

				content = append(content, data)
				filled[index] = true
				index++
			}

			timer.Stop()
		}

	}()

	return res
}

// Float64CombineInPartialOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// Float64CombineInPartialOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is not stopped and will continue but there will be an empty slot in the slice returned.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64CombineInPartialOrder(ctx context.Context, maxItemWait time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]float64, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		closed := make(map[int]bool, 0)

		var sendersClosed int

		for {
			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] || closed[index] {
				index++
				continue
			}

			if sendersClosed >= total {
				res <- content
				return
			}

			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]float64, len(senders))
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					timer.Stop()

					sendersClosed++
					closed[index] = true
					continue
				}

				content[index] = data
				filled[index] = true
				index++
			}

			timer.Stop()
		}

	}()

	return res
}

// Float64CombineInOrder receives a giving stream of content from multiple channels, returning a single channel of a
// two slice type, more so, it will use the maxItemWait duration to give every channel a opportunity for delivery
// data. If the time is passed it will cycle to the next item, till the complete set is retrieved from, unless
// the context timout expires and causes a complete stop of operation. CombineInOrder guarantees that the data
// retrieved will be stored in order of passed in channels. If any of the channels is nil, then the returned
// channel will be closed.
// CombineInOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func Float64CombineInOrder(ctx context.Context, maxItemWait time.Duration, senders ...chan float64) chan []float64 {
	res := make(chan []float64, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]float64, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] {
				index++
				continue
			}

			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]float64, len(senders))
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total {
				case true:
					index = 0
				case false:
					index++
				}
			case data, ok := <-senders[index]:
				if !ok {
					close(res)
					timer.Stop()
					return
				}

				content[index] = data
				filled[index] = true
				index++
			}

			timer.Stop()
		}

	}()

	return res
}

// Float64Distributor delivers messages to subscription channels which it manages internal,
// ensuring every subscriber gets the delivered message, it guarantees that every subscribe will get the chance to receive a
// messsage unless it takes more than a giving duration of time, and if passing that duration that the operation
// to deliver will be cancelled, this ensures that we do not leak goroutines nor
// have eternal channel blocks.
type Float64Distributor struct {
	running             int64
	messages            chan float64
	close               chan struct{}
	clear               chan struct{}
	subscribers         []chan float64
	newSub              chan chan float64
	sendWaitBeforeAbort time.Duration
}

// NewFloat64Disributor returns a new instance of a Float64Distributor.
func NewFloat64Disributor(buffer int, sendWaitBeforeAbort time.Duration) *Float64Distributor {
	if sendWaitBeforeAbort <= 0 {
		sendWaitBeforeAbort = defaultSendWithBeforeAbort
	}

	return &Float64Distributor{
		clear:               make(chan struct{}, 0),
		close:               make(chan struct{}, 0),
		subscribers:         make([]chan float64, 0),
		newSub:              make(chan chan float64, 0),
		messages:            make(chan float64, buffer),
		sendWaitBeforeAbort: sendWaitBeforeAbort,
	}
}

// PublishDeadline sends the message into the distributor to be delivered to all subscribers if it has not
// passed the provided deadline.
func (d *Float64Distributor) PublishDeadline(message float64, dur time.Duration) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
		return
	case d.messages <- message:
		return
	}
}

// Publish sends the message into the distributor to be delivered to all subscribers.
func (d *Float64Distributor) Publish(message float64) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.messages <- message
}

// Subscribe adds the channel into the distributor subscription lists.
func (d *Float64Distributor) Subscribe(sub chan float64) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.newSub <- sub
}

// Clear removes all subscribers from the distributor.
func (d *Float64Distributor) Clear() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.clear <- struct{}{}
}

// Stop halts internal delivery behaviour of the distributor.
func (d *Float64Distributor) Stop() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.close <- struct{}{}
}

// Start initializes the distributor to deliver messages to subscribers.
func (d *Float64Distributor) Start() {
	if atomic.LoadInt64(&d.running) != 0 {
		return
	}

	atomic.AddInt64(&d.running, 1)
	go d.manage()
}

// manage implements necessary logic to manage message delivery and
// subscriber adding
func (d *Float64Distributor) manage() {
	defer atomic.AddInt64(&d.running, -1)

	for {
		select {
		case <-d.clear:
			d.subscribers = nil

		case newSub, ok := <-d.newSub:
			if !ok {
				return
			}

			d.subscribers = append(d.subscribers, newSub)
		case message, ok := <-d.messages:
			if !ok {
				return
			}

			for _, sub := range d.subscribers {
				go func(c chan float64) {
					tick := time.NewTimer(d.sendWaitBeforeAbort)
					defer tick.Stop()

					select {
					case sub <- message:
						return
					case <-tick.C:
						return
					}
				}(sub)
			}
		case <-d.close:
			return
		}
	}
}

// MonoFloat64Service defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve only float64 values from a single source.
// It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type MonoFloat64Service interface {
	// Send will return a channel which will allow reading from the Service it till it is closed.
	Send() (<-chan float64, error)

	// Receive will take the channel, which will be written into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Receive(<-chan float64) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() chan struct{}

	// Errors returns a channel which signals services to know whether the Service is still servicing
	// request.
	Errors() chan error

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}

// Float64Service defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve float64 values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type Float64Service interface {
	// Send will return a channel which will allow reading from the Service it till it is closed.
	Send(string) (<-chan float64, error)

	// Receive will take the channel, which will be written into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Receive(string, <-chan float64) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() chan struct{}

	// Errors returns a channel which signals services to know whether the Service is still servicing
	// request.
	Errors() chan error

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}
