package json

import (
	"encoding/json"
	"time"

	"github.com/influx6/dime/services"
)

// UInt64BecomeBytesWithJSON implements an adapter for uint64 using the JSON encoder.
// Any content that fails the JSON.Marshal will be ignored.
func UInt64BecomeBytesWithJSON(ctx services.CancelContext, in <-chan uint64) <-chan []byte {
	output := make(chan []byte, 0)

	go func() {
		t := time.NewTimer(1 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				t.Reset(1 * time.Second)
				continue
			case indata, ok := <-in:
				if !ok {
					return
				}

				outdata, err := json.Marshal(indata)
				if err != nil {
					continue
				}

				output <- outdata
			}
		}
	}()

	return output
}
