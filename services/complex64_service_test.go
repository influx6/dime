package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/tests"
)

func TestComplex64Collect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64Collect(ctx, 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- 1.0:

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64Collect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- 1.0:

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64Collect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			for i := 3; i > 0; i-- {

				incoming <- 2.0

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

func TestComplex64PartialCollect(t *testing.T) {
	t.Logf("When all data is received before 3 second")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64PartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- 1.0:

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64PartialCollect(ctx, 10*time.Millisecond, incoming)

		go func() {
			for i := 20; i > 0; i-- {
				select {
				case <-ctx.Done():
					return

				case incoming <- 1.0:

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64PartialCollect(context.Background(), 10*time.Millisecond, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- 2.0

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

func TestComplex64Mutate(t *testing.T) {
	t.Logf("When data is mutated but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64Mutate(ctx, 2*time.Millisecond, func(item complex64) complex64 { return item }, incoming)

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64Mutate(ctx, 10*time.Millisecond, func(item complex64) complex64 {
			return item
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- 2.0

			}
		}()

		_, ok := <-outgoing
		if !ok {
			tests.Failed("Should have recieved item as value but got %t", ok)
		}
		tests.Passed("Should have recieved false as value")

		_, ok = <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestComplex64View(t *testing.T) {
	t.Logf("When data is View but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64View(ctx, 2*time.Millisecond, func(item complex64) {}, incoming)

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

		incoming := make(chan complex64, 0)
		inview := make(chan complex64, 0)

		outgoing := services.Complex64View(ctx, 10*time.Millisecond, func(item complex64) {
			inview <- item
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- 2.0

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

func TestComplex64Filter(t *testing.T) {
	t.Logf("When data is filtered but not received due to context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64Filter(ctx, 2*time.Millisecond, func(item complex64) bool {
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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64Filter(ctx, 10*time.Millisecond, func(item complex64) bool {
			return true
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 1; i > 0; i-- {

				incoming <- 2.0

			}
		}()

		_, ok := <-outgoing
		if !ok {
			tests.Failed("Should have recieved only 1 item as value but got %t", ok)
		}
		tests.Passed("Should have recieved false as value")

		_, ok = <-outgoing
		if ok {
			tests.Failed("Should have recieved close signal")
		}
		tests.Passed("Should have recieved close signal")
	}
}

func TestComplex64CollectUntil(t *testing.T) {
	t.Logf("When data is collected until condition is met except when context expiration on receive")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64CollectUntil(ctx, 2*time.Millisecond, func(item []complex64) bool { return true }, incoming)

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

		incoming := make(chan complex64, 0)
		outgoing := services.Complex64CollectUntil(ctx, 10*time.Millisecond, func(items []complex64) bool {
			return len(items) == 2
		}, incoming)

		go func() {
			defer close(incoming)

			for i := 3; i > 0; i-- {

				incoming <- 2.0

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

func TestComplex64MergeWithoutOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64MergeWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)
		defer close(incoming)

		incoming2 := make(chan complex64, 0)
		defer close(incoming2)

		incoming3 := make(chan complex64, 0)
		defer close(incoming3)

		outgoing := services.Complex64MergeWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestComplex64MergeInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64MergeInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)
		defer close(incoming)

		incoming2 := make(chan complex64, 0)
		defer close(incoming2)

		incoming3 := make(chan complex64, 0)
		defer close(incoming3)

		outgoing := services.Complex64MergeInOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 9 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 9 item slice")
	}
}

func TestComplex64CombinePartiallyWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels without provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64CombinePartiallyWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)
		defer close(incoming)

		incoming2 := make(chan complex64, 0)
		defer close(incoming2)

		incoming3 := make(chan complex64, 0)
		defer close(incoming3)

		outgoing := services.Complex64CombinePartiallyWithoutOrder(ctx, 1*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestComplex64CombineWithoutOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in incoming order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64CombineWithoutOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)

		incoming2 := make(chan complex64, 0)

		incoming3 := make(chan complex64, 0)

		outgoing := services.Complex64CombineWithoutOrder(ctx, 10*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)
			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			defer close(incoming2)
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			defer close(incoming3)
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestComplex64CombineInOrder(t *testing.T) {
	t.Logf("When data is merged from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64CombineInOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)
		defer close(incoming)

		incoming2 := make(chan complex64, 0)
		defer close(incoming2)

		incoming3 := make(chan complex64, 0)
		defer close(incoming3)

		outgoing := services.Complex64CombineInOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")
	}
}

func TestComplex64CombineInPartialOrder(t *testing.T) {
	t.Logf("When data is combined from multiple channels in provided order but context expires so nothing is received")
	{

		ctx, cancl := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancl()

		incoming := make(chan complex64, 0)
		defer close(incoming)

		outgoing := services.Complex64CombineInPartialOrder(ctx, 2*time.Millisecond, incoming)

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

		incoming := make(chan complex64, 0)

		incoming2 := make(chan complex64, 0)
		defer close(incoming2)

		incoming3 := make(chan complex64, 0)
		defer close(incoming3)

		outgoing := services.Complex64CombineInPartialOrder(ctx, 2*time.Millisecond, incoming, incoming2, incoming3)

		go func() {
			defer close(incoming)

			time.Sleep(3 * time.Millisecond)

			incoming <- 2.0

		}()

		go func() {
			time.Sleep(1 * time.Millisecond)

			incoming2 <- 2.0

		}()

		go func() {
			time.Sleep(2 * time.Millisecond)

			incoming3 <- 2.0

		}()

		received := <-outgoing
		if received == nil {
			tests.Failed("Should have recieved 3 item slice but got %d item slice", len(received))
		}
		tests.Passed("Should have recieved 3 item slice")

	}
}

func TestComplex64Distributor(t *testing.T) {
	dist := services.NewComplex64Distributor(0, 1*time.Second)

	incoming := make(chan complex64, 1)
	incoming2 := make(chan complex64, 1)
	incoming3 := make(chan complex64, 1)

	dist.Subscribe(incoming)
	dist.Subscribe(incoming2)
	dist.Subscribe(incoming3)

	defer dist.Stop()

	dist.Publish(

		2.0,
	)

	select {
	case <-time.After(15 * time.Millisecond):
		tests.Failed("Should have received a published item before expiration")
	case initial := <-incoming:
		tests.Passed("Should have received a matching data on first channel")

		if !isComplex64Equal(initial, <-incoming2) {
			tests.Failed("Should have received a matching data on second channel")
		}
		tests.Passed("Should have received a matching data on second channel")

		if !isComplex64Equal(initial, <-incoming3) {
			tests.Failed("Should have received a matching data on third channel")
		}
		tests.Passed("Should have received a matching data on third channel")
	}

	dist.Stop()

	dist.PublishDeadline(

		2.0,

		1*time.Millisecond,
	)

	if len(incoming) != 0 || len(incoming2) != 0 || len(incoming3) != 0 {
		tests.Failed("Should not have received any items after publisher is stopped")
	}
	tests.Passed("Should not have received any items after publisher is stopped")
}

func isComplex64Equal(item1, item2 complex64) bool {

	if item1 != item2 {
		return false
	}

	return true
}

func isComplex64EqualSlice(item1 []complex64, item2 []complex64) bool {
	if len(item1) != len(item2) {
		return false
	}

	for index, item := range item1 {
		if !isComplex64Equal(item, item2[index]) {
			return false
		}
	}

	return true
}
