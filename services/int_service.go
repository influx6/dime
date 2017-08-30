package services

import (
	"sync/atomic"
	"time"
)

//go:generate moz generate-file -fromFile ./int_service.go -toDir ./impl/int

// IntDataWriterFunc defines a function type which recieves a value to be written and returns true/false
// if the operation succeeded.
type IntDataWriterFunc func(int) bool

// IntFromByteAdapter defines a function that that will take a channel of bytes and return a channel of int.
type IntFromByteAdapter func(CancelContext, <-chan []byte) <-chan int

// IntToByteAdapter defines a function that that will take a channel of bytes and return a channel of int.
type IntToByteAdapter func(CancelContext, <-chan int) <-chan []byte

// IntPartialCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed or the context expires returning whatever was collected, and closing the returning channel.
// This function does not guarantee complete data, because if the context expires, what is already gathered even if incomplete is returned.
func IntPartialCollect(ctx CancelContext, waitTime time.Duration, in <-chan int) <-chan []int {
	res := make(chan []int, 0)

	go func() {
		var buffer []int

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

// IntCollect defines a function which returns a channel where the items of the incoming channel
// are buffered until the channel is closed, nothing will be returned if the channel given is not closed  or the context expires.
// Once done, returning channel is closed.
// This function guarantees complete data.
func IntCollect(ctx CancelContext, waitTime time.Duration, in <-chan int) <-chan []int {
	res := make(chan []int, 0)

	go func() {
		var buffer []int

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

// IntMutate defines a function which returns a channel where the items of the incoming channel
// are mutated based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func IntMutate(ctx CancelContext, waitTime time.Duration, mutateFn func(int) int, in <-chan int) <-chan int {
	res := make(chan int, 0)

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

// IntView defines a function which returns a channel where the items of the incoming channel
// are provided to function after delivry to output channel, till the provided channel is closed.
// This guarantees that whatever the function sees is something which has being delivered to the output
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func IntView(ctx CancelContext, waitTime time.Duration, viewFn func(int), in <-chan int) <-chan int {
	res := make(chan int, 0)

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

// IntSink defines a function which returns a channel, where the items of the returned channel
// are to be writting to the incoming channel, till the returned channel is closed which will lead to the
// closure of the incoming channed.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func IntSink(ctx CancelContext, waitTime time.Duration, in chan<- int) chan<- int {
	res := make(chan int, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(in)
				return

			case data, ok := <-res:
				if !ok {
					close(in)
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

// IntWriterFuncToWithin defines a function which recieves a CancelContext, max time.Duration and write channel,
// to return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired or the maxAcceptance duration was not exceeded
// on every call.
// The returned function returns true/false to signal success write of value.
func IntWriterFuncToWithin(ctx CancelContext, acceptanceMaxWait time.Duration, in chan<- int) IntDataWriterFunc {
	return func(val int) bool {
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

// IntWriterFuncTo defines a function which recieves a CancelContext and write channel, to
// return a function that writes new incoming values and guarantee that the provided value will be delivered to
// the provided wrie channel while the CancelContext has not expired on every call.
// The returned function returns true/false to signal success write of value.
func IntWriterFuncTo(ctx CancelContext, in chan<- int) IntDataWriterFunc {
	return func(val int) bool {
		select {
		case <-ctx.Done():
			return false
		case in <- val:
			return true
		}
	}
}

// IntSinkFilter defines a function which returns a channel where the items of the returned channel
// are provided to function which filters incoming values and allows only acceptable values, which is delivered
// to the incoming channel, till the returned channel is closed by the user and will lead to the closure of the
// incoming channel as well.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func IntSinkFilter(ctx CancelContext, waitTime time.Duration, filterFn func(int) bool, in chan<- int) chan<- int {
	res := make(chan int, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(in)
				return

			case data, ok := <-res:
				if !ok {
					close(in)
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

// IntSinkMutate defines a function which returns a channel where the items of the returned channel
// are provided to function which mutates and returns a new value then which is  delivered to the incoming channel,
// till the returned channel is closed by the user and will lead to the closure of the incoming channel as well.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func IntSinkMutate(ctx CancelContext, waitTime time.Duration, mutateFn func(int) int, in chan<- int) chan<- int {
	res := make(chan int, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(in)
				return

			case data, ok := <-res:
				if !ok {
					close(in)
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

// IntSinkView defines a function which returns a channel where the items of the returned channel
// are provided to function after delivry to incoming channel, till the returned channel is closed by the user
// and will lead to the closure of the incoming channel as well.
// This guarantees that whatever the function sees is something which has being written to the incoming channel
// and was accepted. Also, receiving function must be careful not to modify incoming value or do so cautiously.
// If the given channel is closed or if the context expires, the incoming channel is closed as well.
// This function guarantees complete data.
// Extreme care must be taking by the user of the returned channel to do a select on with the CancelContext has he/she/it
// sends data into the returned channel to ensure that it is closed and stopped once context has expired by it's Done()
// method.
func IntSinkView(ctx CancelContext, waitTime time.Duration, viewFn func(int), in chan<- int) chan<- int {
	res := make(chan int, 0)

	go func() {
		t := time.NewTimer(waitTime)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				close(in)
				return

			case data, ok := <-res:
				if !ok {
					close(in)
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

// IntFilter defines a function which returns a channel where the items of the incoming channel
// are filtered based on a function, till the provided channel is closed.
// If the given channel is closed or if the context expires, the returning channel is closed as well.
// This function guarantees complete data.
func IntFilter(ctx CancelContext, waitTime time.Duration, filterFn func(int) bool, in <-chan int) <-chan int {
	res := make(chan int, 0)

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

// IntCollectUntil defines a function which returns a channel where the items of the incoming channel
// are buffered until the data matches a given requirement provided through a function. If the function returns true
// then currently buffered data is returned and a new buffer is created. This is useful for batch collection based on
// specific criteria. If the channel is closed before the criteria is met, what data is left is sent down the returned channel,
// closing that channel. If the context expires then data gathered is returned and returning channel is closed.
// This function guarantees some data to be delivered.
func IntCollectUntil(ctx CancelContext, waitTime time.Duration, condition func([]int) bool, in <-chan int) <-chan []int {
	res := make(chan []int, 0)

	go func() {
		var buffer []int

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

// IntMergeWithoutOrder merges the incoming data from the int into a single stream of int,
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
func IntMergeWithoutOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

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
		filledContent := make(map[int]int, 0)

		for {
			if len(filled) == total {
				var content []int

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

// IntMergeInOrder merges the incoming data from the int into a single stream of int,
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
func IntMergeInOrder(ctx CancelContext, maxWaitTime time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

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
		filledContent := make(map[int]int, 0)

		for {
			if len(filled) == total {
				var content []int

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

// IntCombineParitallyWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// IntCombineParitallyWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func IntCombinePartiallyWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]int, 0)

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
				content = make([]int, len(senders))
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

// IntCombineWithoutOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// IntCombineWithoutOrder makes the following guarantees:
// 1. Items will be received in any order received from the channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is stopped and the returned channel is closed.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect data in any order for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func IntCombineWithoutOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]int, 0)

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(content) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]int, len(senders))
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

// IntCombineInPartialOrder receives a giving stream of content from multiple channels, returning a single channel of a
// 2d slice, it sequentially tries to recieve data from each sender within a provided time duration, else skipping until it's next turn.
// This ensures every sender has adquate time to receive and reducing long blocked waits for a specific sender, more so, this reduces the
// overhead of managing multiple go-routined receving channels which are prone to goroutine ophaning or memory leaks.
// IntCombineInPartialOrder makes the following guarantees:
// 1. Items will be received in order and return in order of channels provided.
// 2. Returned channel will never return incomplete data where a channels return value is missing from its index.
// 3. If any channel is closed then all operation is not stopped and will continue but there will be an empty slot in the slice returned.
// 4. If the context expires before items are complete then returned channel is closed.
// 5. It will continously collect order data for all channels until any of the above conditions are broken.
// 6. All channel data are collected once for the receiving scope, i.e a channels data will not be received twice into the return slice
//    but all channels will have a single data slot for a partial data collection session.
// 7. Will continue to gather data from provided channels until all are closed or the context has expired.
// 8. If any of the senders is nil then the returned channel will be closed, has this leaves things in an unstable state.
func IntCombineInPartialOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]int, len(senders))

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
				content = make([]int, len(senders))
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

// IntCombineInOrder receives a giving stream of content from multiple channels, returning a single channel of a
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
func IntCombineInOrder(ctx CancelContext, maxItemWait time.Duration, senders ...<-chan int) <-chan []int {
	res := make(chan []int, 0)

	for _, elem := range senders {
		if elem == nil {
			close(res)
			return res
		}
	}

	go func() {
		content := make([]int, len(senders))

		var index int

		total := len(senders)
		filled := make(map[int]bool, 0)

		for {
			if len(filled) == total {
				res <- content
				index = 0
				filled = make(map[int]bool, 0)
				content = make([]int, len(senders))
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

// IntDistributor delivers messages to subscription channels which it manages internal,
// ensuring every subscriber gets the delivered message, it guarantees that every subscribe will get the chance to receive a
// messsage unless it takes more than a giving duration of time, and if passing that duration that the operation
// to deliver will be cancelled, this ensures that we do not leak goroutines nor
// have eternal channel blocks.
type IntDistributor struct {
	running             int64
	messages            chan int
	closer              chan struct{}
	clear               chan struct{}
	subscribers         []chan<- int
	newSub              chan chan<- int
	sendWaitBeforeAbort time.Duration
}

// NewIntDistributor returns a new instance of a IntDistributor.
func NewIntDistributor(buffer int, sendWaitBeforeAbort time.Duration) *IntDistributor {
	if sendWaitBeforeAbort <= 0 {
		sendWaitBeforeAbort = defaultSendWithBeforeAbort
	}

	return &IntDistributor{
		clear:               make(chan struct{}, 0),
		closer:              make(chan struct{}, 0),
		subscribers:         make([]chan<- int, 0),
		newSub:              make(chan chan<- int, 0),
		messages:            make(chan int, buffer),
		sendWaitBeforeAbort: sendWaitBeforeAbort,
	}
}

// PublishDeadline sends the message into the distributor to be delivered to all subscribers if it has not
// passed the provided deadline.
func (d *IntDistributor) PublishDeadline(message int, dur time.Duration) {
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
func (d *IntDistributor) Publish(message int) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.messages <- message
}

// Subscribe adds the channel into the distributor subscription lists.
func (d *IntDistributor) Subscribe(sub chan<- int) {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.newSub <- sub
}

// Clear removes all subscribers from the distributor.
func (d *IntDistributor) Clear() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.clear <- struct{}{}
}

// Stop halts internal delivery behaviour of the distributor.
func (d *IntDistributor) Stop() {
	if atomic.LoadInt64(&d.running) == 0 {
		return
	}

	d.closer <- struct{}{}
}

// Start initializes the distributor to deliver messages to subscribers.
func (d *IntDistributor) Start() {
	if atomic.LoadInt64(&d.running) != 0 {
		return
	}

	atomic.AddInt64(&d.running, 1)
	go d.manage()
}

// manage implements necessary logic to manage message delivery and
// subscriber adding
func (d *IntDistributor) manage() {
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
				go func(c chan<- int) {
					tick := time.NewTimer(d.sendWaitBeforeAbort)
					defer tick.Stop()

					select {
					case c <- message:
						return
					case <-tick.C:
						return
					}
				}(sub)
			}
		case <-d.closer:
			return
		}
	}
}

// MonoIntService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve only int values from a single source.
// It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type MonoIntService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read() (<-chan int, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(<-chan int) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}

// IntService defines a interface for underline systems which want to communicate like
// in a stream using channels to send/recieve int values. It allows different services to create adapters to
// transform data coming in and out from the Service.
// Auto-Generated using the moz code-generator https://github.com/influx6/moz.
// @iface
type IntService interface {
	// ReadErrors will return a channel which will allow reading errors from the Service until it it is closed.
	ReadErrors() <-chan error

	// Read will return a channel which will allow reading from the Service until it it is closed.
	Read(string) (<-chan int, error)

	// Receive will take the channel, which will be writing into the Service for it's internal processing
	// and the Service will continue to read form the channel till the channel is closed.
	// Useful for collating/collecting services.
	Write(string, <-chan int) error

	// Done defines a signal to other pending services to know whether the Service is still servicing
	// request.
	Done() <-chan struct{}

	// Service defines a function to be called to stop the Service internal operation and to close
	// all read/write operations.
	Stop() error
}
