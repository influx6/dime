package impl

import (
	"errors"
	"time"
)

// timeouts
var (
	writerWaitDuration       = 100 * time.Millisecond
	readerCloseCheckDuration = 10 * time.Millisecond
	errorWriteAcceptTimeout  = 10 * time.Millisecond
)

// errors
var (
	ErrNotSupported = errors.New("Not Supported")
)
