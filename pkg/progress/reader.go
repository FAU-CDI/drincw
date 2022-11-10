package progress

import (
	"fmt"
	"io"
	"time"

	"github.com/dustin/go-humanize"
)

type Reader struct {
	io.Reader
	Bytes int64

	lastFlush time.Time
	Progress  io.Writer
}

func (cr *Reader) Read(bytes []byte) (int, error) {
	count, err := cr.Reader.Read(bytes)
	cr.Bytes += int64(count)
	cr.Flush(false)
	return count, err
}

const flushInterval = time.Second / 30

func (cr *Reader) Flush(force bool) {
	if force || time.Since(cr.lastFlush) > flushInterval {
		cr.lastFlush = time.Now()
		fmt.Fprintf(cr.Progress, "\r Read %s", humanize.Bytes(uint64(cr.Bytes)))
	}
}
