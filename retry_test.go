package retry

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNumberOfRetriesError(t *testing.T) {
	var sum int64
	Retry(func() error {
		atomic.AddInt64(&sum, 1)
		return errors.Errorf("DUMMY")
	},
		3,
		nil,
		time.Millisecond*50)
	assert.Equal(t, int64(3), sum)
}

func TestNumberOfRetriesPanic(t *testing.T) {
	var sum int64
	Retry(func() error {
		atomic.AddInt64(&sum, 1)
		return errors.Errorf("DUMMY")
	},
		3,
		nil,
		time.Millisecond*50)
	assert.Equal(t, int64(3), sum)
}

func TestNumberOfRetriesNoError(t *testing.T) {
	var sum int64
	Retry(func() error {
		atomic.AddInt64(&sum, 1)
		return nil
	},
		3,
		nil,
		time.Millisecond*50)
	assert.Equal(t, int64(1), sum)
}

func TestOnError(t *testing.T) {
	var sum int64
	Retry(func() error {
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		3,
		func(error) { atomic.AddInt64(&sum, 1) },
		time.Millisecond*50)
	assert.Equal(t, int64(6), sum)
}

func TestOnError2(t *testing.T) {
	var sum int64
	Retry(func() error {
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		3000,
		func(error) { atomic.AddInt64(&sum, 1) },
		time.Microsecond)
	assert.Equal(t, int64(6000), sum)
}

func TestOnError3(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var sum int64
	Retry(func() error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		3,
		func(error) { cancel() },
		time.Millisecond*50)
	<-ctx.Done()
	assert.Equal(t, int64(1), sum)
}

func ExampleTry() {
	Try(func() error {
		fmt.Println("done")
		return nil
	})

	// Output:
	// done
}

func ExampleTry_error() {
	if err := Try(func() error {
		return errors.New("FAILED")
	}); err != nil {
		fmt.Println(err)
	}

	// Output:
	// FAILED
}

func ExampleRetry() {
	Retry(func() error {
		fmt.Println(1)
		return errors.Errorf("FAILED")
	},
		3, nil,
		time.Millisecond*50)

	// Output:
	// 1
	// 1
	// 1
}

func ExampleRetry_error() {
	Retry(func() error {
		return errors.Errorf("FAILED")
	},
		3,
		func(err error) { fmt.Println(err) },
		time.Millisecond*50)

	// Output:
	// FAILED
	// FAILED
	// FAILED
}

func ExampleRetry_panic() {
	Retry(func() error {
		panic(errors.Errorf("FAILED"))
	},
		3, func(err error) { fmt.Println(err) },
		time.Millisecond*50)

	// Output:
	// FAILED
	// FAILED
	// FAILED
}

func ExampleRetry_period() {
	startedAt := time.Now()
	Retry(func() error {
		return errors.Errorf("FAILED")
	},
		3, func(err error) { fmt.Println(time.Since(startedAt).Round(time.Millisecond * 5)) },
		time.Millisecond*50)

	// Output:
	// 0s
	// 50ms
	// 100ms
}

func ExampleRetry_scheduler() {
	// run every 50 millisecond, for 3 times:
	var cnt int64
	reschedule := errors.Errorf("re-schedule")
	Retry(func() error {
		atomic.AddInt64(&cnt, 1)
		return reschedule
	},
		3, func(err error) {
			if err == reschedule {
				return
			}
			fmt.Println(err)
		},
		time.Millisecond*50)

	fmt.Println(cnt)

	// Output:
	// 3
}
