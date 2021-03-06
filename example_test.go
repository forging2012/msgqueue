package msgqueue_test

import (
	"errors"
	"fmt"
	"math"
	"time"

	"golang.org/x/time/rate"

	"github.com/go-msgqueue/msgqueue"
	"github.com/go-msgqueue/msgqueue/memqueue"
)

func timeSince(start time.Time) time.Duration {
	secs := float64(time.Since(start)) / float64(time.Second)
	return time.Duration(math.Floor(secs)) * time.Second
}

func timeSinceCeil(start time.Time) time.Duration {
	secs := float64(time.Since(start)) / float64(time.Second)
	return time.Duration(math.Ceil(secs)) * time.Second
}

func Example_retryOnError() {
	start := time.Now()
	q := memqueue.NewQueue(&msgqueue.Options{
		Handler: func() error {
			fmt.Println("retried in", timeSince(start))
			return errors.New("fake error")
		},
		RetryLimit: 3,
		MinBackoff: time.Second,
	})

	q.Call()

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: retried in 0s
	// retried in 1s
	// retried in 3s
}

func Example_messageDelay() {
	start := time.Now()
	q := memqueue.NewQueue(&msgqueue.Options{
		Handler: func() {
			fmt.Println("processed with delay", timeSince(start))
		},
	})

	msg := msgqueue.NewMessage()
	msg.Delay = time.Second
	q.Add(msg)

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: processed with delay 1s
}

func Example_rateLimit() {
	start := time.Now()
	q := memqueue.NewQueue(&msgqueue.Options{
		Handler: func() {
			fmt.Println(timeSinceCeil(start))
		},
		Redis:     redisRing(),
		RateLimit: rate.Every(time.Second),
	})

	for i := 0; i < 5; i++ {
		q.Call()
	}

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: 1s
	// 1s
	// 2s
	// 3s
	// 4s
}

func Example_once() {
	q := memqueue.NewQueue(&msgqueue.Options{
		Handler: func(name string) {
			fmt.Println("hello", name)
		},
		Redis:     redisRing(),
		RateLimit: rate.Every(time.Second),
	})

	for i := 0; i < 10; i++ {
		// Call once in a second.
		q.CallOnce(time.Second, "world")
	}

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: hello world
}

func Example_maxWorkers() {
	start := time.Now()
	q := memqueue.NewQueue(&msgqueue.Options{
		Handler: func() {
			fmt.Println(timeSince(start))
			time.Sleep(time.Second)
		},
		Redis:       redisRing(),
		WorkerLimit: 1,
	})

	for i := 0; i < 3; i++ {
		q.Call()
	}

	// Wait for all messages to be processed.
	_ = q.Close()

	// Output: 0s
	// 1s
	// 2s
}
