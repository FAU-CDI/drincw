package perf_test

import (
	"fmt"
	"runtime"
	"time"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

// An example of capturing performance metrics
func ExampleNow() {
	metrics := perf.Now()

	// some fancy and slow task
	{
		var stuff [10000]int
		defer runtime.KeepAlive(stuff)
		time.Sleep(1 * time.Second)
	}

	// print out performance metrics
	fmt.Println(perf.Since(metrics))
}

func ExampleDiff() {
	// Diff holds both the amount of time an operation took,
	// the number of bytes consumed, and the total number of allocated objects.
	diff := perf.Diff{
		Time:    15 * time.Second,
		Bytes:   100,
		Objects: 100,
	}
	fmt.Println(diff)
	// Output: 15s, 100 B, 100 objects
}