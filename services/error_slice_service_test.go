package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/tests"
)

func TestErrorSliceCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceCollect(ctx, 10*time.Millisecond, incoming)

		// Awat
		go func() {
			defer close(incoming)

			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []error{errors.New(fmt.Sprintf("%q", "item"))}:

					continue
				}
			}
		}()

		received, ok := <-outgoing
		if !ok {
			tests.Failed("Should have fully received all items before 100ms")
		}
		tests.Passed("Should have fully received all items before 100ms")

		if received == nil {
			tests.Failed("Should have received 20 items but got %d", len(received))
		}
		tests.Passed("Should have received 20 items.")
	}

	t.Logf("When all data is not received after 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []error{errors.New(fmt.Sprintf("%q", "item"))}:

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

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			for i := 3; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

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

func TestErrorSlicePartialCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []error{errors.New(fmt.Sprintf("%q", "item"))}:

					continue
				}
			}
		}()

		received, ok := <-outgoing
		if !ok {
			tests.Failed("Should have fully received all items before 100ms")
		}
		tests.Passed("Should have fully received all items before 100ms")

		if received == nil {
			tests.Failed("Should have received 20 items but got %d", len(received))
		}
		tests.Passed("Should have received 20 items.")
	}

	t.Logf("When all data is received after 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []error{errors.New(fmt.Sprintf("%q", "item"))}:

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

		if received == nil {
			tests.Failed("Should have received 2 items but got %d", len(received))
		}
		tests.Passed("Should have received %d items.", len(received))
	}

	t.Logf("When sending channel is closed we should still get collected items")
	{
		ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSlicePartialCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

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

			if received == nil {
				tests.Failed("Should have received 3 items but got %d", len(received))
			}
			tests.Passed("Should have received %d items.", len(received))
		}
	}
}

func TestErrorSliceMutate(t *testing.T) {
	t.Logf("When data is mutated but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceMutate(ctx, 2*time.Millisecond, func(item []error) []error { return item }, incoming)

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

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceMutate(ctx, 10*time.Millisecond, func(item []error) []error {
			if len(item) > 2 {
				return item[0:2]
			} else {
				return item
			}
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

			}
		}()

		received1 := <-outgoing
		if received1 == nil {
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

func TestErrorSliceView(t *testing.T) {
	t.Logf("When data is View but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceView(ctx, 2*time.Millisecond, func(item []error) {}, incoming)

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

		incoming := make(chan []error, 0)
		inview := make(chan []error, 0)

		outgoing := services.ErrorSliceView(ctx, 10*time.Millisecond, func(item []error) {
			inview <- item
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

			}
		}()

		_, ok := <-outgoing
		if !ok {
			tests.Failed("Should have recieved only 1 item as value but got %t", ok)
		}
		tests.Passed("Should have recieved 1 item")

		_, ok = <-inview
		if !ok {
			tests.Failed("Should have recieved only 1 item as value but got %t", ok)
		}
		tests.Passed("Should have recieved 1 item")

		_, ok = <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved 1 item")
	}
}

func TestErrorSliceFilter(t *testing.T) {
	t.Logf("When data is filtered but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceFilter(ctx, 2*time.Millisecond, func(item []error) bool {
			return true
		}, incoming)
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

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceFilter(ctx, 10*time.Millisecond, func(item []error) bool {
			return true
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

			}
		}()

		received1 := <-outgoing
		if received1 == nil {
			tests.Failed("Should have recieved only 3 as value but got %t", received1)
		}
		tests.Passed("Should have recieved false as value")

		_, ok := <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestErrorSliceCollectUntil(t *testing.T) {
	t.Logf("When data is collected until condition is met except when context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceCollectUntil(ctx, 2*time.Millisecond, func(item [][]error) bool { return true }, incoming)

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

		incoming := make(chan []error, 0)
		outgoing := services.ErrorSliceCollectUntil(ctx, 10*time.Millisecond, func(items [][]error) bool {
			return len(items) == 2
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

			}
		}()

		received1 := <-outgoing
		if received1 == nil {
			tests.Failed("Should have recieved 2 item slice but got %d item slice", len(received1))
		}
		tests.Passed("Should have recieved 2 item slice")

		received2 := <-outgoing
		if received2 == nil {
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

func TestErrorSliceMergeWithoutOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceMergeWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)
		defer close(incoming)

		incoming2 := make(chan []error, 0)
		defer close(incoming2)

		incoming3 := make(chan []error, 0)
		defer close(incoming3)

		outgoing := services.ErrorSliceMergeWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestErrorSliceMergeInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceMergeInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)
		defer close(incoming)

		incoming2 := make(chan []error, 0)
		defer close(incoming2)

		incoming3 := make(chan []error, 0)
		defer close(incoming3)

		outgoing := services.ErrorSliceMergeInOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestErrorSliceCombinePartiallyWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels without provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceCombinePartiallyWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)
		defer close(incoming)

		incoming2 := make(chan []error, 0)
		defer close(incoming2)

		incoming3 := make(chan []error, 0)
		defer close(incoming3)

		outgoing := services.ErrorSliceCombinePartiallyWithoutOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestErrorSliceCombineWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceCombineWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)

		incoming2 := make(chan []error, 0)

		incoming3 := make(chan []error, 0)

		outgoing := services.ErrorSliceCombineWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)
			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			defer close(incoming2)
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			defer close(incoming3)
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestErrorSliceCombineInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceCombineInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)
		defer close(incoming)

		incoming2 := make(chan []error, 0)
		defer close(incoming2)

		incoming3 := make(chan []error, 0)
		defer close(incoming3)

		outgoing := services.ErrorSliceCombineInOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")
	}
}

func TestErrorSliceCombineInPartialOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []error, 0)
		defer close(incoming)

		outgoing := services.ErrorSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []error, 0)

		incoming2 := make(chan []error, 0)
		defer close(incoming2)

		incoming3 := make(chan []error, 0)
		defer close(incoming3)

		outgoing := services.ErrorSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)

			time.Sleep(3 * time.Millisecond)

			incoming <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestErrorSliceDistributor(t *testing.T) {
	dist := services.NewErrorSliceDistributor(0, 1*time.Second)

	incoming := make(chan []error, 1)
	incoming2 := make(chan []error, 1)
	incoming3 := make(chan []error, 1)

	dist.Subscribe(incoming)
	dist.Subscribe(incoming2)
	dist.Subscribe(incoming3)

	defer dist.Stop()

	dist.Publish(

		[]error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))},
	)

	select {
	case <-time.After(15 * time.Millisecond):
		tests.Failed("Should have received a published item before expiration")
	case initial := <-incoming:
		tests.Passed("Should have received a matching data on first channel")

		if !isErrorSliceEqualSlice(initial, <-incoming2) {
			tests.Failed("Should have received a matching data on second channel")
		}
		tests.Passed("Should have received a matching data on second channel")

		if !isErrorSliceEqualSlice(initial, <-incoming3) {
			tests.Failed("Should have received a matching data on third channel")
		}
		tests.Passed("Should have received a matching data on third channel")
	}

	dist.Stop()

	dist.PublishDeadline(

		[]error{errors.New(fmt.Sprintf("%q", "item")), errors.New(fmt.Sprintf("%q", "item-2")), errors.New(fmt.Sprintf("%q", 3))},

		1*time.Millisecond,
	)

	if len(incoming) != 0 || len(incoming2) != 0 || len(incoming3) != 0 {
		tests.Failed("Should not have received any items after publisher is stopped")
	}
	tests.Passed("Should not have received any items after publisher is stopped")
}

func isErrorSliceEqualSlice(item1 []error, item2 []error) bool {

	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		if item2[index] != item {
			return false
		}
	}

	return true
}

func isErrorSliceEqualDoubleSlice(item1 [][]error, item2 [][]error) bool {
	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		if !isErrorSliceEqualSlice(item, item2[index]) {
			return false
		}
	}

	return true
}
