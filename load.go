package drincw

import (
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strings"
)

func LoadPathbuilderXML(path string) (*XMLPathbuilder, error) {
	bytes, err := readURLOrPath(path)
	if err != nil {
		return nil, err
	}

	pb := &XMLPathbuilder{}
	if err := xml.Unmarshal(bytes, pb); err != nil {
		return nil, err
	}

	return pb, err
}

// readURLOrPath opens a URL or a file
func readURLOrPath(path string) ([]byte, error) {
	if !(strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")) {
		return os.ReadFile(path)
	}

	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
