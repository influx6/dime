package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/tests"
)

func TestMapSliceCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceCollect(ctx, 10*time.Millisecond, incoming)

		// Awat
		go func() {
			defer close(incoming)

			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []map[string]string{map[string]string{"day": "monday"}}:

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []map[string]string{map[string]string{"day": "monday"}}:

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			for i := 3; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSlicePartialCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []map[string]string{map[string]string{"day": "monday"}}:

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSlicePartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- []map[string]string{map[string]string{"day": "monday"}}:

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSlicePartialCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSliceMutate(t *testing.T) {
	t.Logf("When data is mutated but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceMutate(ctx, 2*time.Millisecond, func(item []map[string]string) []map[string]string { return item }, incoming)

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceMutate(ctx, 10*time.Millisecond, func(item []map[string]string) []map[string]string {
			if len(item) > 2 {
				return item[0:2]
			} else {
				return item
			}
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSliceView(t *testing.T) {
	t.Logf("When data is View but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceView(ctx, 2*time.Millisecond, func(item []map[string]string) {}, incoming)

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

		incoming := make(chan []map[string]string, 0)
		inview := make(chan []map[string]string, 0)

		outgoing := services.MapSliceView(ctx, 10*time.Millisecond, func(item []map[string]string) {
			inview <- item
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSliceFilter(t *testing.T) {
	t.Logf("When data is filtered but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceFilter(ctx, 2*time.Millisecond, func(item []map[string]string) bool {
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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceFilter(ctx, 10*time.Millisecond, func(item []map[string]string) bool {
			return true
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSliceCollectUntil(t *testing.T) {
	t.Logf("When data is collected until condition is met except when context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceCollectUntil(ctx, 2*time.Millisecond, func(item [][]map[string]string) bool { return true }, incoming)

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

		incoming := make(chan []map[string]string, 0)
		outgoing := services.MapSliceCollectUntil(ctx, 10*time.Millisecond, func(items [][]map[string]string) bool {
			return len(items) == 2
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- []map[string]string{map[string]string{"day": "monday"}}

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

func TestMapSliceMergeWithoutOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceMergeWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		incoming2 := make(chan []map[string]string, 0)
		defer close(incoming2)

		incoming3 := make(chan []map[string]string, 0)
		defer close(incoming3)

		outgoing := services.MapSliceMergeWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestMapSliceMergeInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceMergeInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		incoming2 := make(chan []map[string]string, 0)
		defer close(incoming2)

		incoming3 := make(chan []map[string]string, 0)
		defer close(incoming3)

		outgoing := services.MapSliceMergeInOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestMapSliceCombinePartiallyWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels without provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceCombinePartiallyWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		incoming2 := make(chan []map[string]string, 0)
		defer close(incoming2)

		incoming3 := make(chan []map[string]string, 0)
		defer close(incoming3)

		outgoing := services.MapSliceCombinePartiallyWithoutOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestMapSliceCombineWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceCombineWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)

		incoming2 := make(chan []map[string]string, 0)

		incoming3 := make(chan []map[string]string, 0)

		outgoing := services.MapSliceCombineWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)
			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			defer close(incoming2)
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			defer close(incoming3)
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestMapSliceCombineInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceCombineInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		incoming2 := make(chan []map[string]string, 0)
		defer close(incoming2)

		incoming3 := make(chan []map[string]string, 0)
		defer close(incoming3)

		outgoing := services.MapSliceCombineInOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")
	}
}

func TestMapSliceCombineInPartialOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan []map[string]string, 0)
		defer close(incoming)

		outgoing := services.MapSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan []map[string]string, 0)

		incoming2 := make(chan []map[string]string, 0)
		defer close(incoming2)

		incoming3 := make(chan []map[string]string, 0)
		defer close(incoming3)

		outgoing := services.MapSliceCombineInPartialOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)

			time.Sleep(3 * time.Millisecond)

			incoming <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- []map[string]string{map[string]string{"day": "monday"}}

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestMapSliceDistributor(t *testing.T) {
	dist := services.NewMapSliceDistributor(0, 1*time.Second)

	incoming := make(chan []map[string]string, 1)
	incoming2 := make(chan []map[string]string, 1)
	incoming3 := make(chan []map[string]string, 1)

	dist.Subscribe(incoming)
	dist.Subscribe(incoming2)
	dist.Subscribe(incoming3)

	defer dist.Stop()

	dist.Publish(

		[]map[string]string{map[string]string{"day": "monday"}},
	)

	select {
	case <-time.After(15 * time.Millisecond):
		tests.Failed("Should have received a published item before expiration")
	case initial := <-incoming:
		tests.Passed("Should have received a matching data on first channel")

		if !isMapSliceEqualSlice(initial, <-incoming2) {
			tests.Failed("Should have received a matching data on second channel")
		}
		tests.Passed("Should have received a matching data on second channel")

		if !isMapSliceEqualSlice(initial, <-incoming3) {
			tests.Failed("Should have received a matching data on third channel")
		}
		tests.Passed("Should have received a matching data on third channel")
	}

	dist.Stop()

	dist.PublishDeadline(

		[]map[string]string{map[string]string{"day": "monday"}},

		1*time.Millisecond,
	)

	if len(incoming) != 0 || len(incoming2) != 0 || len(incoming3) != 0 {
		tests.Failed("Should not have received any items after publisher is stopped")
	}
	tests.Passed("Should not have received any items after publisher is stopped")
}

func isMapEqualForMapSlice(item1, item2 map[string]string) bool {
	if item1 == nil {
		return false
	}

	for index, item := range item1 {
		if item2[index] != item {
			return false
		}
	}

	return true
}

func isMapSliceEqualSlice(item1 []map[string]string, item2 []map[string]string) bool {

	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		citem := item2[index]
		if !isMapEqualForMapSlice(item, citem) {
			return false
		}
	}

	return true
}

func isMapSliceEqualDoubleSlice(item1 [][]map[string]string, item2 [][]map[string]string) bool {
	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		if !isMapSliceEqualSlice(item, item2[index]) {
			return false
		}
	}

	return true
}
