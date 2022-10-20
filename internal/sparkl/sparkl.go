// Package sparkl implements a very primitive graph index
package sparkl

import (
	"io"
	"os"

	"github.com/cayleygraph/quad"
	"github.com/cayleygraph/quad/nquads"
)

// LoadIndex is like ReadIndex, but reads it from the given path
func LoadIndex(path string) (*GraphIndex[string, any], error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return ReadIndex(reader)
}

// ReadIndex creates a new index to be read from nquads
func ReadIndex(r io.Reader) (*GraphIndex[string, any], error) {
	// open a reader
	reader := nquads.NewReader(r, true)
	defer reader.Close()

	// create a new index
	index := &GraphIndex[string, any]{}
	index.Reset()

	// insert stuff into the index
	for {
		q, err := reader.ReadQuad()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		sI, sOK := asURILike(q.Subject.Native())
		pI, pOK := asURILike(q.Predicate.Native())
		if !(sOK && pOK) {
			continue
		}

		o := q.Object.Native()
		oI, oOK := asURILike(o)
		if oOK {
			index.AddTriple(sI, pI, oI)

		} else {
			index.AddData(sI, pI, o)
		}
	}

	// and finalize the index
	index.Finalize()
	return index, nil
}

func asURILike(value any) (uri string, ok bool) {
	switch datum := value.(type) {
	case quad.IRI:
		return string(datum), true
	case quad.BNode:
		return string(datum), true
	default:
		return "", false
	}
}
