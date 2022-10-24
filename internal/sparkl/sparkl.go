// Package sparkl implements a very primitive graph index
package sparkl

import (
	"os"
)

// LoadIndex is like ReadIndex, but reads it from the given path
func LoadIndex(path string, sameAsPredicates []string) (*Index, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ReadNQuads(reader, sameAsPredicates)
}
