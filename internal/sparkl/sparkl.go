// Package sparkl implements a very primitive graph index
package sparkl

import (
	"os"
)

// LoadIndex is like MakeIndex, but reads nquads from the given path
func LoadIndex(path string, predicates Predicates) (*Index, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return MakeIndex(&QuadSource{Reader: reader}, predicates)
}
