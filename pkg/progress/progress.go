// Package progress provides Reader and Writer
package progress

import (
	"fmt"
	"io"
	"time"

	"github.com/dustin/go-humanize"
)

// Reader consistently writes the number of bytes read to Progress.
type Reader struct {
	io.Reader       // Reader to read from
	Bytes     int64 // total number of bytes read (so far)

	FlushInterval time.Duration // minimum time between flushes of the progress
	lastFlush     time.Time     // last time we flushed

	Progress io.Writer // Progress output to write to
}

func (cr *Reader) Read(bytes []byte) (int, error) {
	count, err := cr.Reader.Read(bytes)
	cr.Bytes += int64(count)
	cr.Flush(false)
	return count, err
}

// Flush flushes the progress to Progress
func (cr *Reader) Flush(force bool) {
	if force || time.Since(cr.lastFlush) > cr.FlushInterval {
		cr.lastFlush = time.Now()
		fmt.Fprintf(cr.Progress, "\rRead %s", humanize.Bytes(uint64(cr.Bytes)))
	}
}

// Writer consistently writes the number of bytes written to Progress.
type Writer struct {
	io.Writer       // Writer to write to
	Bytes     int64 // Total number of bytes written

	FlushInterval time.Duration // minimum time between flushes of the progress
	lastFlush     time.Time

	Progress io.Writer // where to write progress to
}

func (cw *Writer) Write(bytes []byte) (int, error) {
	cw.Bytes += int64(len(bytes))
	cw.Flush(false)
	return cw.Writer.Write(bytes)
}

func (cw *Writer) Flush(force bool) {
	if force || time.Since(cw.lastFlush) > cw.FlushInterval {
		cw.lastFlush = time.Now()
		fmt.Fprintf(cw.Progress, "\rWrote %s", humanize.Bytes(uint64(cw.Bytes)))
	}
}

// DefaultFlushInterval is a reasonable default flush interval
const DefaultFlushInterval = time.Second / 30
