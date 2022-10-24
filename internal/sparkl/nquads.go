package sparkl

import (
	"io"

	"github.com/cayleygraph/quad"
	"github.com/cayleygraph/quad/nquads"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type Index = igraph.IGraph[string, any]

const SameAs = "http://www.w3.org/2002/07/owl#sameAs"

// ReadNQuads reads NQuads from the given reader
func ReadNQuads(r io.ReadSeeker, SameAsPredicates []string) (*Index, error) {
	// create a new index
	var index Index
	index.Reset()

	// read the "same as" triples first
	if err := readSameAs(r, &index, SameAsPredicates); err != nil {
		return nil, err
	}

	// and then read all the other data
	if err := readData(r, &index); err != nil {
		return nil, err
	}

	// and finalize the index
	index.Finalize()
	return &index, nil
}

func readSameAs(r io.ReadSeeker, index *Index, sameAsPredicates []string) error {
	if len(sameAsPredicates) == 0 {
		return nil
	}

	sameAss := make(map[string]struct{})
	for _, sameAs := range sameAsPredicates {
		sameAss[sameAs] = struct{}{}
	}

	for q := range scanQuads(r) {
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

func readData(r io.ReadSeeker, index *Index) error {
	for q := range scanQuads(r) {
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

// q represents a scanned quad
type q struct {
	subject   string
	predicate string

	hasDatum bool
	object   string
	datum    any

	err error
}

// scanQuads scans quads from the start of the reader
func scanQuads(r io.ReadSeeker) <-chan q {
	r.Seek(0, io.SeekStart)
	reader := nquads.NewReader(r, true)

	results := make(chan q)

	go func() {
		defer close(results)

		for {
			value, err := reader.ReadQuad()
			if err == io.EOF {
				break
			}
			if err != nil {
				results <- q{err: err}
			}

			sI, sOK := asURILike(value.Subject.Native())
			pI, pOK := asURILike(value.Predicate.Native())
			if !(sOK && pOK) {
				continue
			}

			o := value.Object.Native()
			oI, oOK := asURILike(o)
			if oOK {
				results <- q{
					subject:   sI,
					predicate: pI,
					object:    oI,
				}
			} else {
				results <- q{
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
