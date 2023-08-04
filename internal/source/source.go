// Package source provides ReadAll.
package source

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// ReadAll reads all data from the given source.
//
// Source can be a remote url starting with 'http://' or 'https://', in which case [http.Get] is used to fetch it's content.
// In all other cases, source is assumed to be a local path, which is read with [os.ReadFile].
func ReadAll(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return readURL(source)
	}

	return os.ReadFile(source)
}

// readURL implements fetching all content from the given url
func readURL(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
