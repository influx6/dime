package services

import (
	"sync"
	"sync/atomic"
	"time"
)

//go:generate moz generate-file -fromFile ./interface_service.go -toDir ./impl/interface

// InterfaceDataWriterFunc defines a function type which recieves a value to be written and returns true/false
// if the operation succeeded.
type InterfaceDataWriterFunc func(interface{}) bool

// InterfaceFromByteAdapter defines a function that that will take a channel of bytes and return a channel of interface{}.
type InterfaceFromByteAdapter func(CancelContext, <-chan []byte) <-chan interface{}

// InterfaceToByteAdapter defines a function that that will take a channel of bytes and return a channel of interface{}.
type InterfaceToByteAdapter func(CancelContext, <-chan interface{}) <-chan []byte

// InterfacePartialCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed or the context expires returning whatever was collected, and closing the returning channel.
// This function does not guarantee complete data, because if the context expires, what is already gathered even if incomplete is returned.
func InterfacePartialCollect(ctx CancelContext, waitTime time.Duration, in <-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	go func() {
		var buffer []interface{}

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

// InterfaceCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed, nothing will be returned if the channel given is not closed  or the context expires.
// Once done, returning channel is closed.
// This function guarantees complete data.
func InterfaceCollect(ctx CancelContext, waitTime time.Duration, in <-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	go func() {
		var buffer []interface{}

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

// InterfaceMutate defines a function which returns a channel where the items of the incoming channel
// are mutated based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func InterfaceMutate(ctx CancelContext, waitTime time.Duration, mutateFn func(interface{}) interface{}, in <-chan interface{}) <-chan interface{} {
	res := make(chan interface{}, 0)

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

// InterfaceView defines a function which returns a channel where the items of the incoming channel
// are provided to function after delivry to output channel, till the provided channel is closed.
// This guarantees that whatever the function sees is something which has being delivered to the output
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func InterfaceView(ctx CancelContext, waitTime time.Duration, viewFn func(interface{}), in <-chan interface{}) <-chan interface{} {
	res := make(chan interface{}, 0)

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

				res <- data
				viewFn(data)
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// InterfaceSink defines a function which returns a channel, where the items of the returned channel
// are to be writting to the incoming channel, till the returned channel is closed which will lead to the
// closure of the incoming channed if closeInputAlso flag is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func InterfaceSink(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, in chan<- interface{}) chan<- interface{} {
	res := make(chan interface{}, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				if closeInputAlso {
					close(in)
				}
				return

			case data, ok := <-res:
				if !ok {
					if closeInputAlso {
						close(in)
					}
					return
				}

				in <- data
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// InterfaceWriterFuncToWithin defines a function which recieves a CancelContext, max time.Duration and write channel,
// to return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired or the maxAcceptance duration was not exceeded
// on every call.
// The returned function returns true/false to signal success write of value.
func InterfaceWriterFuncToWithin(ctx CancelContext, acceptanceMaxWait time.Duration, in chan<- interface{}) InterfaceDataWriterFunc {
	return func(val interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		case in <- val:
			return true
		case <-time.After(acceptanceMaxWait):
			return false
		}
	}
}

// InterfaceWriterFuncTo defines a function which recieves a CancelContext and write channel, to
// return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired on every call.
// The returned function returns true/false to signal success write of value.
func InterfaceWriterFuncTo(ctx CancelContext, in chan<- interface{}) InterfaceDataWriterFunc {
	return func(val interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		case in <- val:
			return true
		}
	}
}

// InterfaceSinkFilter defines a function which returns a channel where the items of the returned channel
// are provided to function which filters incoming values and allows only acceptable values, which is delivered
// to the incoming channel, till the returned channel is closed by the user and will lead to the closure of the
// incoming channed if closeInputAlso flag is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func InterfaceSinkFilter(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, filterFn func(interface{}) bool, in chan<- interface{}) chan<- interface{} {
	res := make(chan interface{}, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				if closeInputAlso {
					close(in)
				}
				return

			case data, ok := <-res:
				if !ok {
					if closeInputAlso {
						close(in)
					}
					return
				}

				if !filterFn(data) {
					continue
				}

				in <- data
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// InterfaceSinkMutate defines a function which returns a channel where the items of the returned channel
// are provided to function which mutates and returns a new value then which is  delivered to the incoming channel,
// till the returned channel is closed by the user and will lead to the closure of the incoming channel as well if
// the closeInputAlso flag is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func InterfaceSinkMutate(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, mutateFn func(interface{}) interface{}, in chan<- interface{}) chan<- interface{} {
	res := make(chan interface{}, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				if closeInputAlso {
					close(in)
				}
				return

			case data, ok := <-res:
				if !ok {
					if closeInputAlso {
						close(in)
					}
					return
				}

				in <- mutateFn(data)
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// InterfaceSinkView defines a function which returns a channel where the items of the returned channel
// are provided to function after delivry to incoming channel, till the returned channel is closed by the user
// and will lead to the closure of the incoming channel as well if the closeInputAlso is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func InterfaceSinkView(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, viewFn func(interface{}), in chan<- interface{}) chan<- interface{} {
	res := make(chan interface{}, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				if closeInputAlso {
					close(in)
				}
				return

			case data, ok := <-res:
				if !ok {
					if closeInputAlso {
						close(in)
					}
					return
				}

				in <- data
				viewFn(data)
			case <-t.C:
				t.Reset(waitTime)
				continue
			}
		}
	}()

	return res
}

// InterfaceFilter defines a function which returns a channel where the items of the incoming channel
// are filtered based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func InterfaceFilter(ctx CancelContext, waitTime time.Duration, filterFn func(interface{}) bool, in <-chan interface{}) <-chan interface{} {
	res := make(chan interface{}, 0)

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

// InterfaceCollectUntil defines a function which returns a channel where the items of the incoming channel
// are buffered until the data matches a given requirement provided through a function. If the function returns true
// then currently buffered data is returned and a new buffer is created. This is useful for batch collection based on
// specific criteria. If the channel is closed before the criteria is met, what data is left is sent down the returned channel,
// closing that channel. If the context expires then data gathered is returned and returning channel is closed.
// This function guarantees some data to be delivered.
func InterfaceCollectUntil(ctx CancelContext, waitTime time.Duration, condition func([]interface{}) bool, in <-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	go func() {
		var buffer []interface{}

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
				if !condition(buffer) {
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

// InterfaceMergeWithoutOrder merges the incoming data from the interface{} into a single stream of interface{},
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
func InterfaceMergeWithoutOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		filledContent := make(map[int]interface{}, 0)

		for {
			if len(filled) == total {
				var content []interface{}

				for _, item := range filledContent {
					content = append(content, item)
				}

				res <- content

				index = 0
				filled = make(map[int]bool, 0)
			}

			// if the current index has being filled, shift forward and reattempt loop.
			if ok := filled[index]; ok {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxWaitTime)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				filled[index] = true
				filledContent[index] = data

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}
	}()

	return res
}

// InterfaceMergeInOrder merges the incoming data from the interface{} into a single stream of interface{},
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
func InterfaceMergeInOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		filledContent := make(map[int]interface{}, 0)

		for {
			if len(filled) == total {
				var content []interface{}

				for index := range senders {
					item := filledContent[index]
					content = append(content, item)
				}

				res <- content

				index = 0
				filled = make(map[int]bool, 0)
			}

			// if the current index has being filled, shift forward and reattempt loop.
			if ok := filled[index]; ok {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxWaitTime)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				filled[index] = true
				filledContent[index] = data

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}
	}()

	return res
}

// InterfaceCombineParitallyWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// InterfaceCombineParitallyWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func InterfaceCombinePartiallyWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]interface{}, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		closed := make(map[int]bool, 0)

		var sendersClosed int

		for {
			if sendersClosed >= total {
				res <- content
				return
			}

			if len(content) == total {
				res <- content

				index = 0
				filled = make(map[int]bool, 0)
				content = make([]interface{}, len(senders))
			}

			// if the current index has being filled, shift forward and re-attempt loop.
			if filled[index] || closed[index] {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}

	}()

	return res
}

// InterfaceCombineWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// InterfaceCombineWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func InterfaceCombineWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]interface{}, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(content) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]interface{}, len(senders))
			}

			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}

	}()

	return res
}

// InterfaceCombineInPartialOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// InterfaceCombineInPartialOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is not stopped and will continue but there will be an empty slot in the slice returned.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func InterfaceCombineInPartialOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]interface{}, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)
		closed := make(map[int]bool, 0)

		var sendersClosed int

		for {
			if sendersClosed >= total {
				res <- content
				return
			}

			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]interface{}, len(senders))
			}

			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] || closed[index] {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}

	}()

	return res
}

// InterfaceCombineInOrder receives a giving stream of content from multiple channels, returning a single channel of a
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
func InterfaceCombineInOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan interface{}) <-chan []interface{} {
	res := make(chan []interface{}, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]interface{}, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]interface{}, len(senders))
			}

			// if the current index has being filled, shift forward and reattempt loop.
			if filled[index] {
				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
				continue
			}

			timer := time.NewTimer(maxItemWait)

			select {
			case <-ctx.Done():
				close(res)
				timer.Stop()
				return
			case <-timer.C:
				switch index >= total-1 {
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

				switch index >= total-1 {
				case true:
					index = 0
				case false:
					index++
				}
			}

			timer.Stop()
		}

	}()

	return res
}

// InterfaceDistributor delivers messages to subscription channels which it manages internal,
// ensuring every subscriber gets the delivered message, it guarantees that every subscribe will get the chance to receive a
// messsage unless it takes more than a giving duration of time, and if passing that duration that the operation
// to deliver will be cancelled, this ensures that we do not leak goroutines nor
// have eternal channel blocks.
type InterfaceDistributor struct {
	running     int64
	messages    chan interface{}
	closer      chan struct{}
	subcloser   chan struct{}
	clear       chan struct{}
	receivers   sync.WaitGroup
	subscribers []chan<- interface{}
	newSub      chan chan<- interface{}
	// sendWaitBeforeAbort time.Duration

}

// NewInterfaceDistributor returns a new instance of a InterfaceDistributor.
func NewInterfaceDistributor(buffer int) *InterfaceDistributor {
	dist := &InterfaceDistributor{
		clear:       make(chan struct{}, 0),
		closer:      make(chan struct{}, 0),
		subcloser:   make(chan struct{}, 0),
		subscribers: make([]chan<- interface{}, 0),
		newSub:      make(chan chan<- interface{}, 0),
		messages:    make(chan interface{}, buffer),
	}

	atomic.AddInt64(&dist.running, 1)
	go dist.manage()

	return dist
}

// PublishDeadline sends the message into the distributor to be delivered to all subscribers if it has not
// passed the provided deadline.
func (d *InterfaceDistributor) PublishDeadline(message interface{}, dur time.Duration) {
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
func (d *InterfaceDistributor) Publish(message interface{}) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.messages <- message
}

// Subscribe adds the channel into the distributor subscription lists.
func (d *InterfaceDistributor) Subscribe(sub chan<- interface{}) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.newSub <- sub
}

// Clear removes all subscribers from the distributor.
func (d *InterfaceDistributor) Clear() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.clear <- struct{}{}
}

// Stop halts internal delivery behaviour of the distributor.
func (d *InterfaceDistributor) Stop() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.closer <- struct{}{}

	d.receivers.Wait()

	for _, sub := range d.subscribers {
		close(sub)
	}

}

// manage implements necessary logic to manage message delivery and
// subscriber adding
func (d *InterfaceDistributor) manage() {
	defer atomic.AddInt64(&d.running, -1)

	ticker := time.NewTimer(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ticker.Reset(1 * time.Second)
			continue

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
				d.receivers.Add(1)

				go func(c chan<- interface{}) {
					defer d.receivers.Done()
					select {
					case <-d.subcloser:
						return
					case c <- message:
						return
					}
				}(sub)
			}
		case <-d.closer:
			close(d.subcloser)
			return
		}
	}
}

// MonoInterfaceService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve only interface{} values from a single source.
// It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type MonoInterfaceService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read() (<-chan interface{}, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(<-chan interface{}) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}

// InterfaceService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve interface{} values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type InterfaceService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read(string) (<-chan interface{}, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(string, <-chan interface{}) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}
