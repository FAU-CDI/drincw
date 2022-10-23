package sparkl

import (
	"io"

	"github.com/cayleygraph/quad"
	"github.com/cayleygraph/quad/nquads"
)

// ReadNQuads reads NQuads from the given reader
func ReadNQuads(r io.Reader) (*GraphIndex[string, any], error) {
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
