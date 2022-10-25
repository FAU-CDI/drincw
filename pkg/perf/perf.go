// Package perf captures performance metrics
package perf

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
)

// Snapshot represents metrics at a specific point in time.
type Snapshot struct {
	Time  time.Time
	Bytes uint64
}

func (snapshot Snapshot) String() string {
	return fmt.Sprintf("%s used at %s", humanize.Bytes(snapshot.Bytes), snapshot.Time.Format(time.Stamp))
}

// Sub subtracts the metrics for a difference snapshots
func (s Snapshot) Sub(other Snapshot) Diff {
	return Diff{
		Time: s.Time.Sub(other.Time),
	}
}

// Now returns a snapshot for the current moment
func Now() Snapshot {
	return Snapshot{
		Time:  time.Now(),
		Bytes: measureHeapCount(),
	}
}

// Diff represents the difference between two points in time
type Diff struct {
	Time  time.Duration
	Bytes uint64
}

func (diff Diff) String() string {
	return fmt.Sprintf("%s, %s", diff.Time, humanize.Bytes(diff.Bytes))
}

// Since computes the diff between now, and the previous point in time
func Since(start Snapshot) Diff {
	return Diff{
		Time:  time.Since(start.Time),
		Bytes: measureHeapCount() - start.Bytes,
	}
}

const (
	measureHeapThreshold = 10 * 1024                           // number of bytes to be considered stable time
	measureHeapSleep     = 50 * time.Millisecond               // amount of time to sleep between measuring cyles
	measureMaxCyles      = int(time.Second / measureHeapSleep) // maximal cycles to run
)

// measureHeapCount measures the current use of the heap
func measureHeapCount() uint64 {
	// NOTE(twiesing): This has been vaguely adapted from https://dev.to/vearutop/estimating-memory-footprint-of-dynamic-structures-in-go-2apf

	var stats runtime.MemStats

	var prevHeapUse, currentHeapUse uint64
	var prevGCCount, currentGCCount uint32

	for i := 0; i < measureMaxCyles; i++ {
		runtime.ReadMemStats(&stats)
		currentGCCount = stats.NumGC
		currentHeapUse = stats.HeapInuse

		if prevGCCount != 0 && currentGCCount > prevGCCount && math.Abs(float64(currentHeapUse-prevHeapUse)) < measureHeapThreshold {
			break
		}

		prevHeapUse = currentHeapUse
		prevGCCount = currentGCCount

		time.Sleep(measureHeapSleep)
		runtime.GC()
	}

	return currentHeapUse
}
