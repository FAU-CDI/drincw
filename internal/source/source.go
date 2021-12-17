// Package source reads data from a remote or local source.
package source

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// ReadAll reads all data from src.
// Source may be a local path, or a url that must start with 'http://' or 'https://'.
//
// Local files are read using os.ReadFile.
// Remote URLs are first fetched using http.Get and then passed to io.ReadAll.
func ReadAll(src string) ([]byte, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return readURI(src)
	}
	return os.ReadFile(src)
}

func readURI(uri string) ([]byte, error) {
	res, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
