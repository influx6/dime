package services

import (
	"sync"
	"sync/atomic"
	"time"
)

//go:generate moz generate-file -fromFile ./bool_service.go -toDir ./impl/bool

// BoolDataWriterFunc defines a function type which recieves a value to be written and returns true/false
// if the operation succeeded.
type BoolDataWriterFunc func(bool) bool

// BoolFromByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolFromByteAdapter func(CancelContext, <-chan []byte) <-chan bool

// BoolToByteAdapter defines a function that that will take a channel of bytes and return a channel of bool.
type BoolToByteAdapter func(CancelContext, <-chan bool) <-chan []byte

// BoolPartialCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed or the context expires returning whatever was collected, and closing the returning channel.
// This function does not guarantee complete data, because if the context expires, what is already gathered even if incomplete is returned.
func BoolPartialCollect(ctx CancelContext, waitTime time.Duration, in <-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	go func() {
		var buffer []bool

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

// BoolCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed, nothing will be returned if the channel given is not closed  or the context expires.
// Once done, returning channel is closed.
// This function guarantees complete data.
func BoolCollect(ctx CancelContext, waitTime time.Duration, in <-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	go func() {
		var buffer []bool

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

// BoolMutate defines a function which returns a channel where the items of the incoming channel
// are mutated based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func BoolMutate(ctx CancelContext, waitTime time.Duration, mutateFn func(bool) bool, in <-chan bool) <-chan bool {
	res := make(chan bool, 0)

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

// BoolView defines a function which returns a channel where the items of the incoming channel
// are provided to function after delivry to output channel, till the provided channel is closed.
// This guarantees that whatever the function sees is something which has being delivered to the output
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func BoolView(ctx CancelContext, waitTime time.Duration, viewFn func(bool), in <-chan bool) <-chan bool {
	res := make(chan bool, 0)

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

// BoolSink defines a function which returns a channel, where the items of the returned channel
// are to be writting to the incoming channel, till the returned channel is closed which will lead to the
// closure of the incoming channed if closeInputAlso flag is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func BoolSink(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, in chan<- bool) chan<- bool {
	res := make(chan bool, 0)

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

// BoolWriterFuncToWithin defines a function which recieves a CancelContext, max time.Duration and write channel,
// to return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired or the maxAcceptance duration was not exceeded
// on every call.
// The returned function returns true/false to signal success write of value.
func BoolWriterFuncToWithin(ctx CancelContext, acceptanceMaxWait time.Duration, in chan<- bool) BoolDataWriterFunc {
	return func(val bool) bool {
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

// BoolWriterFuncTo defines a function which recieves a CancelContext and write channel, to
// return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired on every call.
// The returned function returns true/false to signal success write of value.
func BoolWriterFuncTo(ctx CancelContext, in chan<- bool) BoolDataWriterFunc {
	return func(val bool) bool {
		select {
		case <-ctx.Done():
			return false
		case in <- val:
			return true
		}
	}
}

// BoolSinkFilter defines a function which returns a channel where the items of the returned channel
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
func BoolSinkFilter(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, filterFn func(bool) bool, in chan<- bool) chan<- bool {
	res := make(chan bool, 0)

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

// BoolSinkMutate defines a function which returns a channel where the items of the returned channel
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
func BoolSinkMutate(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, mutateFn func(bool) bool, in chan<- bool) chan<- bool {
	res := make(chan bool, 0)

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

// BoolSinkView defines a function which returns a channel where the items of the returned channel
// are provided to function after delivry to incoming channel, till the returned channel is closed by the user
// and will lead to the closure of the incoming channel as well if the closeInputAlso is true.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func BoolSinkView(ctx CancelContext, closeInputAlso bool, waitTime time.Duration, viewFn func(bool), in chan<- bool) chan<- bool {
	res := make(chan bool, 0)

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

// BoolFilter defines a function which returns a channel where the items of the incoming channel
// are filtered based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func BoolFilter(ctx CancelContext, waitTime time.Duration, filterFn func(bool) bool, in <-chan bool) <-chan bool {
	res := make(chan bool, 0)

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

// BoolCollectUntil defines a function which returns a channel where the items of the incoming channel
// are buffered until the data matches a given requirement provided through a function. If the function returns true
// then currently buffered data is returned and a new buffer is created. This is useful for batch collection based on
// specific criteria. If the channel is closed before the criteria is met, what data is left is sent down the returned channel,
// closing that channel. If the context expires then data gathered is returned and returning channel is closed.
// This function guarantees some data to be delivered.
func BoolCollectUntil(ctx CancelContext, waitTime time.Duration, condition func([]bool) bool, in <-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	go func() {
		var buffer []bool

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

// BoolMergeWithoutOrder merges the incoming data from the bool into a single stream of bool,
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
func BoolMergeWithoutOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

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
		filledContent := make(map[int]bool, 0)

		for {
			if len(filled) == total {
				var content []bool

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

// BoolMergeInOrder merges the incoming data from the bool into a single stream of bool,
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
func BoolMergeInOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

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
		filledContent := make(map[int]bool, 0)

		for {
			if len(filled) == total {
				var content []bool

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

// BoolCombineParitallyWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// BoolCombineParitallyWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func BoolCombinePartiallyWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]bool, 0)

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
				content = make([]bool, len(senders))
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

// BoolCombineWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// BoolCombineWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func BoolCombineWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]bool, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(content) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]bool, len(senders))
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

// BoolCombineInPartialOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// BoolCombineInPartialOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is not stopped and will continue but there will be an empty slot in the slice returned.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func BoolCombineInPartialOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]bool, len(senders))

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
				content = make([]bool, len(senders))
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

// BoolCombineInOrder receives a giving stream of content from multiple channels, returning a single channel of a
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
func BoolCombineInOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan bool) <-chan []bool {
	res := make(chan []bool, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]bool, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]bool, len(senders))
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

// BoolDistributor delivers messages to subscription channels which it manages internal,
// ensuring every subscriber gets the delivered message, it guarantees that every subscribe will get the chance to receive a
// messsage unless it takes more than a giving duration of time, and if passing that duration that the operation
// to deliver will be cancelled, this ensures that we do not leak goroutines nor
// have eternal channel blocks.
type BoolDistributor struct {
	running     int64
	messages    chan bool
	closer      chan struct{}
	subcloser   chan struct{}
	clear       chan struct{}
	receivers   sync.WaitGroup
	subscribers []chan<- bool
	newSub      chan chan<- bool
	// sendWaitBeforeAbort time.Duration

}

// NewBoolDistributor returns a new instance of a BoolDistributor.
func NewBoolDistributor(buffer int) *BoolDistributor {
	dist := &BoolDistributor{
		clear:       make(chan struct{}, 0),
		closer:      make(chan struct{}, 0),
		subcloser:   make(chan struct{}, 0),
		subscribers: make([]chan<- bool, 0),
		newSub:      make(chan chan<- bool, 0),
		messages:    make(chan bool, buffer),
	}

	atomic.AddInt64(&dist.running, 1)
	go dist.manage()

	return dist
}

// PublishDeadline sends the message into the distributor to be delivered to all subscribers if it has not
// passed the provided deadline.
func (d *BoolDistributor) PublishDeadline(message bool, dur time.Duration) {
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
func (d *BoolDistributor) Publish(message bool) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.messages <- message
}

// Subscribe adds the channel into the distributor subscription lists.
func (d *BoolDistributor) Subscribe(sub chan<- bool) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.newSub <- sub
}

// Clear removes all subscribers from the distributor.
func (d *BoolDistributor) Clear() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.clear <- struct{}{}
}

// Stop halts internal delivery behaviour of the distributor.
func (d *BoolDistributor) Stop() {
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
func (d *BoolDistributor) manage() {
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

				go func(c chan<- bool) {
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

// MonoBoolService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve only bool values from a single source.
// It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type MonoBoolService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read() (<-chan bool, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(<-chan bool) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}

// BoolService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve bool values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type BoolService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read(string) (<-chan bool, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(string, <-chan bool) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}
