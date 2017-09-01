package functions_test

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/influx6/dime/functions"

	"github.com/influx6/dime/services"
	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/metrics/sentries/stdout"
	"github.com/influx6/faux/tests"
)

var (
	events = metrics.New(stdout.Stdout{})
)

func TestByteFunction(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	var in, out bytes.Buffer
	proc := func(ctx services.CancelContext, mesgs <-chan []byte, replies chan<- []byte, errReplies <-chan error) {
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errReplies:
				if !ok {
					continue
				}

				tests.Errored("Should have not received error: %+q.", err)
			case msg, ok := <-mesgs:
				if !ok {
					// We are required to handle close signal ourselves
					close(replies)
					return
				}

				replies <- []byte(fmt.Sprintf("Hello %s", msg))
			}
		}
	}

	metricProc := functions.WrapWithMetric("sit", events, proc)

	ctx, cn := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cn()

	in.Write([]byte("Alex"))

	go func() {
		defer wg.Done()

		if err := functions.LaunchReaderWriterFunction(ctx, 100, &in, &out, metricProc); err != nil {
			tests.Failed("Should have successfully launched code into opertion: %+q", err)
		}
		tests.Passed("Should have successfully launched code into opertion")
	}()

	wg.Wait()

	fmt.Printf("Out: %+q\n", out.Bytes())
}
