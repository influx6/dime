//
//
//
package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/tests"
)

func TestStringSliceCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceCollect(ctx, 10*time.Millisecond, incoming)

		// Awat
		go func() {
			defer close(incoming)

			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []string{fmt.Sprintf("%q", "monday")}:

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []string{fmt.Sprintf("%q", "monday")}:

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			for i := 3; i > 0; i-- {

				incoming <- []string{fmt.Sprintf("%q", "monday")}

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

func TestStringSlicePartialCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		outgoing := services.StringSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []string{fmt.Sprintf("%q", "monday")}:

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []string{fmt.Sprintf("%q", "monday")}:

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSlicePartialCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []string{fmt.Sprintf("%q", "monday")}

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

func TestStringSliceMutate(t *testing.T) {
	t.Logf("When data is mutated but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceMutate(ctx, 2*time.Millisecond, func(item []string) []string { return item }, incoming)

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceMutate(ctx, 10*time.Millisecond, func(item []string) []string {
			if len(item) > 2 {
				return item[0:2]
			} else {
				return item
			}
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []string{fmt.Sprintf("%q", "monday")}

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

func TestStringSliceFilter(t *testing.T) {
	t.Logf("When data is filtered but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceFilter(ctx, 2*time.Millisecond, func(item []string) bool {
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

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceFilter(ctx, 10*time.Millisecond, func(item []string) bool {
			return true
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []string{fmt.Sprintf("%q", "monday")}

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

func TestStringSliceCollectUntil(t *testing.T) {
	t.Logf("When data is collected until condition is met except when context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceCollectUntil(ctx, 2*time.Millisecond, func(item [][]string) bool { return true }, incoming)

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

		incoming := make(chan []string, 0)
		outgoing := services.StringSliceCollectUntil(ctx, 10*time.Millisecond, func(items [][]string) bool {
			return len(items) == 2
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []string{fmt.Sprintf("%q", "monday")}

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

func TestStringSliceMergeWithoutOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceMergeWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)
		defer close(incoming)

		incoming2 := make(chan []string, 0)
		defer close(incoming2)

		incoming3 := make(chan []string, 0)
		defer close(incoming3)

		outgoing := services.StringSliceMergeWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestStringSliceMergeInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceMergeInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)
		defer close(incoming)

		incoming2 := make(chan []string, 0)
		defer close(incoming2)

		incoming3 := make(chan []string, 0)
		defer close(incoming3)

		outgoing := services.StringSliceMergeInOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestStringSliceCombinePartiallyWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels without provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceCombinePartiallyWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)
		defer close(incoming)

		incoming2 := make(chan []string, 0)
		defer close(incoming2)

		incoming3 := make(chan []string, 0)
		defer close(incoming3)

		outgoing := services.StringSliceCombinePartiallyWithoutOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestStringSliceCombineWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceCombineWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)

		incoming2 := make(chan []string, 0)

		incoming3 := make(chan []string, 0)

		outgoing := services.StringSliceCombineWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)
			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			defer close(incoming2)
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			defer close(incoming3)
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestStringSliceCombineInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceCombineInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)
		defer close(incoming)

		incoming2 := make(chan []string, 0)
		defer close(incoming2)

		incoming3 := make(chan []string, 0)
		defer close(incoming3)

		outgoing := services.StringSliceCombineInOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")
	}
}

func TestStringSliceCombineInPartialOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []string, 0)
		defer close(incoming)

		outgoing := services.StringSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []string, 0)

		incoming2 := make(chan []string, 0)
		defer close(incoming2)

		incoming3 := make(chan []string, 0)
		defer close(incoming3)

		outgoing := services.StringSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)

			time.Sleep(3 * time.Millisecond)

			incoming <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []string{fmt.Sprintf("%q", "monday")}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestStringSliceDistributor(t *testing.T) {
	dist := services.NewStringSliceDisributor(0, 1*time.Second)
	dist.Start()

	incoming := make(chan []string, 1)
	incoming2 := make(chan []string, 1)
	incoming3 := make(chan []string, 1)

	dist.Subscribe(incoming)
	dist.Subscribe(incoming2)
	dist.Subscribe(incoming3)

	dist.Publish(

		[]string{fmt.Sprintf("%q", "monday")},
	)

	select {
	case <-time.After(15 * time.Millisecond):
		tests.Failed("Should have received a published item before expiration")
	case initial := <-incoming:
		tests.Passed("Should have received a matching data on first channel")

		if !isStringSliceEqualSlice(initial, <-incoming2) {
			tests.Failed("Should have received a matching data on second channel")
		}
		tests.Passed("Should have received a matching data on second channel")

		if !isStringSliceEqualSlice(initial, <-incoming3) {
			tests.Failed("Should have received a matching data on third channel")
		}
		tests.Passed("Should have received a matching data on third channel")
	}

	dist.Stop()

	dist.PublishDeadline(

		[]string{fmt.Sprintf("%q", "monday")},

		1*time.Millisecond,
	)

	if len(incoming) != 0 || len(incoming2) != 0 || len(incoming3) != 0 {
		tests.Failed("Should not have received any items after publisher is stopped")
	}
	tests.Passed("Should not have received any items after publisher is stopped")
}

func isStringSliceEqualSlice(item1 []string, item2 []string) bool {

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

func isStringSliceEqualDoubleSlice(item1 [][]string, item2 [][]string) bool {
	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		if !isStringSliceEqualSlice(item, item2[index]) {
			return false
		}
	}

	return true
}
