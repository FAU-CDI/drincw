package progress

import (
	"fmt"
	"io"
	"time"

	"github.com/dustin/go-humanize"
)

type Writer struct {
	io.Writer
	Bytes int64

	lastFlush time.Time
	Progress  io.Writer
}

func (cw *Writer) Write(bytes []byte) (int, error) {
	cw.Bytes += int64(len(bytes))
	fmt.Fprintf(cw.Progress, "\r Wrote %s", humanize.Bytes(uint64(cw.Bytes)))
	return cw.Writer.Write(bytes)
}

func (cw *Writer) Flush(force bool) {
	if force || time.Since(cw.lastFlush) > flushInterval {
		cw.lastFlush = time.Now()
		fmt.Fprintf(cw.Progress, "\r Read %s", humanize.Bytes(uint64(cw.Bytes)))
	}
}
