package sparkl

import (
	"io"

	"github.com/cayleygraph/quad"
	"github.com/cayleygraph/quad/nquads"
)

type ReaderSeeker interface {
	io.Reader
	io.Seeker
}

const SameAs = "http://www.w3.org/2002/07/owl#sameAs"

// ReadNQuads reads NQuads from the given reader
func ReadNQuads(r ReaderSeeker, SameAsPredicates []string) (*GraphIndex[string, any], error) {
	// create a new index
	index := &GraphIndex[string, any]{}
	index.Reset()

	if err := readSameAs(r, index, SameAsPredicates); err != nil {
		return nil, err
	}
	if err := readData(r, index); err != nil {
		return nil, err
	}

	// and finalize the index
	index.Finalize()
	return index, nil
}

func readSameAs(r ReaderSeeker, index *GraphIndex[string, any], SameAsPredicates []string) error {
	if len(SameAsPredicates) == 0 {
		return nil
	}

	sameAss := make(map[string]struct{})
	for _, sameAs := range SameAsPredicates {
		sameAss[sameAs] = struct{}{}
	}

	for q := range readQuads(r) {
		switch {
		case q.err != nil:
			return q.err
		case !q.hasDatum:
			if _, ok := sameAss[q.predicate]; ok {
				index.Identify(q.subject, q.object)
			}
		}
	}
	return nil
}

func readData(r ReaderSeeker, index *GraphIndex[string, any]) error {
	for q := range readQuads(r) {
		switch {
		case q.err != nil:
			return q.err
		case q.hasDatum:
			index.AddData(q.subject, q.predicate, q.datum)
		case !q.hasDatum:
			index.AddTriple(q.subject, q.predicate, q.object)
		}
	}
	return nil
}

type scanned struct {
	subject   string
	predicate string

	hasDatum bool
	object   string
	datum    any

	err error
}

func readQuads(r ReaderSeeker) <-chan scanned {
	r.Seek(0, io.SeekStart)
	reader := nquads.NewReader(r, true)

	results := make(chan scanned)

	go func() {
		defer close(results)

		for {
			q, err := reader.ReadQuad()
			if err == io.EOF {
				break
			}
			if err != nil {
				results <- scanned{err: err}
			}

			sI, sOK := asURILike(q.Subject.Native())
			pI, pOK := asURILike(q.Predicate.Native())
			if !(sOK && pOK) {
				continue
			}

			o := q.Object.Native()
			oI, oOK := asURILike(o)
			if oOK {
				results <- scanned{
					subject:   sI,
					predicate: pI,
					object:    oI,
				}
			} else {
				results <- scanned{
					subject:   sI,
					predicate: pI,
					hasDatum:  true,
					datum:     o,
				}
			}
		}
	}()
	return results
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
