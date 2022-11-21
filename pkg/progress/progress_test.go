package progress

import (
	"fmt"
	"io"
	"strings"
)

func ExampleReader() {
	source := strings.NewReader("hello world")
	var progress strings.Builder

	reader := &Reader{
		Reader: source,

		FlushInterval: 0,
		Progress:      &progress,
	}
	reader.Read([]byte("hello"))
	reader.Read([]byte(" world"))

	// replace all the '\r's with '\n's for testing
	fmt.Println(strings.ReplaceAll(progress.String(), "\r", "\n"))

	// Output: Read 5 B
	// Read 11 B
}

func ExampleWriter() {
	var progress strings.Builder

	writer := &Writer{
		Writer: io.Discard,

		FlushInterval: 0,
		Progress:      &progress,
	}
	writer.Write([]byte("hello"))
	writer.Write([]byte(" world"))

	// replace all the '\r's with '\n's for testing
	fmt.Println(strings.ReplaceAll(progress.String(), "\r", "\n"))

	// Output: Wrote 5 B
	// Wrote 11 B
}
