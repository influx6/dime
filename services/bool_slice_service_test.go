package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/tests"
)

func TestBoolSliceCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSliceCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case incoming <- []bool{(i%2 == 0)}:
					continue
				}
			}
		}()

		received, ok := <-outgoing
		if !ok {
			tests.Failed("Should have fully received all items before 100ms")
		}
		tests.Passed("Should have fully received all items before 100ms")

		if len(received) != 20 {
			tests.Failed("Should have received 20 items but got %d", len(received))
		}
		tests.Passed("Should have received 20 items.")
	}

	t.Logf("When all data is not received after 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSliceCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case incoming <- []bool{(i%2 == 0)}:
					time.Sleep(50 * time.Millisecond)
					continue
				}
			}
		}()

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have not received any items after 100ms")
		}
		tests.Passed("Should have not received any items after 100ms")
	}

	t.Logf("When sending channel is not closed we should not get collected items")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSliceCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			for i := 3; i > 0; i-- {
				incoming <- []bool{(i%2 == 0)}
			}
		}()

		select {
		case <-ctx.Done():
			tests.Passed("Should not have received items after context closed")
		case <-outgoing:
			tests.Failed("Should not have received items after context closed")
		}
	}
}

func TestBoolSlicePartialCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case incoming <- []bool{(i%2 == 0)}:
					continue
				}
			}
		}()

		received, ok := <-outgoing
		if !ok {
			tests.Failed("Should have fully received all items before 100ms")
		}
		tests.Passed("Should have fully received all items before 100ms")

		if len(received) != 20 {
			tests.Failed("Should have received 20 items but got %d", len(received))
		}
		tests.Passed("Should have received 20 items.")
	}

	t.Logf("When all data is received after 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case incoming <- []bool{(i%2 == 0)}:
					time.Sleep(60 * time.Millisecond)
					continue
				}
			}
		}()

		received, ok := <-outgoing
		if !ok {
			tests.Failed("Should have fully received only 2 items after 100ms")
		}
		tests.Passed("Should have fully received only 2 items after 100ms")

		if len(received) != 2 {
			tests.Failed("Should have received 2 items but got %d", len(received))
		}
		tests.Passed("Should have received %d items.", len(received))
	}

	t.Logf("When sending channel is closed we should still get collected items")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSlicePartialCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {
				incoming <- []bool{(i%2 == 0)}
			}
		}()

		select {
		case <-ctx.Done():
			tests.Failed("Should have received items before context closed")
		case received, ok := <-outgoing:
			if !ok {
				tests.Failed("Should have fully received only 2 items after 100ms")
			}
			tests.Passed("Should have fully received only 2 items after 100ms")

			if len(received) != 3 {
				tests.Failed("Should have received 3 items but got %d", len(received))
			}
			tests.Passed("Should have received %d items.", len(received))
		}
	}
}

func TestBoolSliceMutate(t *testing.T) {
	t.Logf("When data is mutated but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		defer close(incoming)

		outgoing := services.BoolSliceMutate(ctx, 2*time.Millisecond, func(item []bool) []bool { return item }, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is mutated on each receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSliceMutate(ctx, 10*time.Millisecond, func(item []bool) []bool {
			return item[0:2]
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {
				incoming <- []bool{(i%2 == 0), (i*2%2 == 0), (i*4%2 == 0)}
			}
		}()

		received1 := <-outgoing
		if len(received1) != 2 {
			tests.Failed("Should have recieved false as value but got %t", received1)
		}
		tests.Passed("Should have recieved false as value")

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestBoolSliceFilter(t *testing.T) {
	t.Logf("When data is filtered but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		defer close(incoming)

		outgoing := services.BoolSliceFilter(ctx, 2*time.Millisecond, func(item []bool) bool { return true }, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is filtered on each receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan []bool, 0)
		outgoing := services.BoolSliceFilter(ctx, 10*time.Millisecond, func(item []bool) bool {
			return item[0] == false
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 2; i > 0; i-- {
				incoming <- []bool{(i%2 == 0)}
			}
		}()

		received1 := <-outgoing
		if len(received1) != 1 {
			tests.Failed("Should have recieved only 1 as value but got %t", received1)
		}
		tests.Passed("Should have recieved false as value")

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestBoolSliceCollectUntil(t *testing.T) {
	t.Logf("When data is collected until condition is met except when context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolCollectUntil(ctx, 2*time.Millisecond, func(item []bool) bool { return true }, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is collected until condition is met")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		outgoing := services.BoolCollectUntil(ctx, 10*time.Millisecond, func(items []bool) bool {
			return len(items) == 2
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {
				incoming <- (i%2 == 0)
			}
		}()

		received1 := <-outgoing
		if len(received1) != 2 {
			tests.Failed("Should have recieved 2 item slice but got %d item slice", len(received1))
		}
		tests.Passed("Should have recieved 2 item slice")

		received2 := <-outgoing
		if len(received2) != 1 {
			tests.Failed("Should have recieved 1 item slice but got %d item slice", len(received2))
		}
		tests.Passed("Should have recieved 1 item slice")

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestBoolSliceMergeWithoutOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolMergeWithoutOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is merged from multiple channels in data incoming order")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		incoming2 := make(chan bool, 0)
		defer close(incoming2)

		incoming3 := make(chan bool, 0)
		defer close(incoming3)

		outgoing := services.BoolMergeWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != false && received[1] != true && received[2] != true {
			tests.Failed("Should have recieved items in matching time.sleep order")
		}
		tests.Passed("Should have recieved items in matching time.sleep order")
	}
}

func TestBoolSliceMergeInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolMergeInOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is merged from multiple channels in provided channel order")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		incoming2 := make(chan bool, 0)
		defer close(incoming2)

		incoming3 := make(chan bool, 0)
		defer close(incoming3)

		outgoing := services.BoolMergeInOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != true && received[1] != true && received[2] != false {
			tests.Failed("Should have recieved items in matching provided order")
		}
		tests.Passed("Should have recieved items in matching provided order")
	}
}

func TestBoolSliceCombinePartiallyWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels without provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolCombinePartiallyWithoutOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is combined from multiple channels without channel order")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		incoming2 := make(chan bool, 0)
		defer close(incoming2)

		incoming3 := make(chan bool, 0)
		defer close(incoming3)

		outgoing := services.BoolCombinePartiallyWithoutOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != true && received[1] != true && received[2] != false {
			tests.Failed("Should have recieved items in matching provided order")
		}
		tests.Passed("Should have recieved items in matching provided order")
	}
}

func TestBoolSliceCombineWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolCombineWithoutOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is merged from multiple channels in data incoming order")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)

		incoming2 := make(chan bool, 0)

		incoming3 := make(chan bool, 0)

		outgoing := services.BoolCombineWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)
			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			defer close(incoming2)
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			defer close(incoming3)
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != false && received[1] != true && received[2] != true {
			tests.Failed("Should have recieved items in matching time.sleep order")
		}
		tests.Passed("Should have recieved items in matching time.sleep order")
	}
}

func TestBoolSliceCombineInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolCombineInOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is merged from multiple channels in provided channel order")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		incoming2 := make(chan bool, 0)
		defer close(incoming2)

		incoming3 := make(chan bool, 0)
		defer close(incoming3)

		outgoing := services.BoolCombineInOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != true && received[1] != true && received[2] != false {
			tests.Failed("Should have recieved items in matching provided order")
		}
		tests.Passed("Should have recieved items in matching provided order")
	}
}

func TestBoolSliceCombineInPartialOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)
		defer close(incoming)

		outgoing := services.BoolCombineInPartialOrder(ctx, 2*time.Millisecond, incoming)

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal due to context expiration")
		}
		tests.Passed("Should have recieved close signal due to context expiration")
	}

	t.Logf("When data is combined from multiple channels in provided channel order")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan bool, 0)

		incoming2 := make(chan bool, 0)
		defer close(incoming2)

		incoming3 := make(chan bool, 0)
		defer close(incoming3)

		outgoing := services.BoolCombineInPartialOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)

			time.Sleep(3 * time.Millisecond)
			incoming <- true
		}()

		go func() {
			time.Sleep(1 * time.Millisecond)
			incoming2 <- false
		}()

		go func() {
			time.Sleep(2 * time.Millisecond)
			incoming3 <- true
		}()

		received := <-outgoing
		if len(received) != 3 {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

		if received[0] != true && received[1] != true && received[2] != false {
			tests.Failed("Should have recieved items in matching provided order")
		}
		tests.Passed("Should have recieved items in matching provided order")
	}
}
