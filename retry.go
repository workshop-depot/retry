// Package retry helps retrying functions in Go, certain number of times or
// infinitely - Also can be used as a schaduler.
package retry

import (
	"time"
)

type recovered struct {
	e interface{}
}

func (r *recovered) Error() string         { return "RECOVERED, UNKNOWN ERROR; CALL CausedBy() interface{}" }
func (r *recovered) CausedBy() interface{} { return r.e }

// Try tries to run a function and recovers from a panic, in case
// one happens, and returns the error, if there are any.
func Try(f func() error) (errRun error) {
	defer func() {
		if e := recover(); e != nil {
			errRun = &recovered{e: e}
		}
	}()
	return f()
}

// Retry retries running a function, numberOfRetries times.
// If numberOfRetries < 0, it runs it forever as long as there are
// any errors. If there are no errors, it will return. If
// numberOfRetries > 1, it will sleep between two attemps,
// the default period is 5 seconds.
func Retry(
	f func() error,
	numberOfRetries int,
	onError func(error),
	period ...time.Duration) {
	p := time.Second * 5
	if len(period) > 0 && period[0] > 0 {
		p = period[0]
	}
	for numberOfRetries != 0 {
		if numberOfRetries > 0 {
			numberOfRetries--
		}
		if err := Try(f); err != nil {
			if onError != nil {
				onError(err)
			}
			if numberOfRetries != 0 {
				time.Sleep(p)
			}
		} else {
			break
		}
	}
}
