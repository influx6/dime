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

	proc := functions.WrapWithMetric("sit", events, func(ctx services.CancelContext, mesgs <-chan []byte, replies chan<- []byte, errReplies <-chan error) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errReplies:
				if !ok {
					return
				}

				tests.Failed("Should have not received error: %+q.", err)
			case msg, ok := <-mesgs:
				if !ok {
					return
				}

				replies <- []byte(fmt.Sprintf("Hello %q\n", msg))
			case <-ticker.C:
				// do nothing
			}
		}
	})

	ctx, cn := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cn()

	go func() {
		defer wg.Done()
		if err := functions.LaunchReaderWriterFunction(ctx, &in, &out, proc); err != nil {
			tests.Failed("Should have successfully launched code into opertion: %+q", err)
		}
		tests.Passed("Should have successfully launched code into opertion")
	}()

	in.Write([]byte("Alex\r\n"))

	wg.Wait()

	fmt.Printf("Out: %+q\n", out.Bytes())
}
