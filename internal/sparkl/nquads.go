package sparkl

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type Index = igraph.IGraph[string, any]

const SameAs = "http://www.w3.org/2002/07/owl#sameAs"
const InverseOf = "http://www.w3.org/2002/07/owl#inverseOf"

// Predicates are predicates with special meaning
type Predicates struct {
	SameAs    []string
	InverseOf []string
}

func MakeIndex(source Source, predicates Predicates) (*Index, error) {
	// create a new index
	var index Index
	index.Reset()

	// read the "same as" triples first
	if err := indexSameAs(source, &index, predicates.SameAs); err != nil {
		return nil, err
	}

	// read the "inverse" triples next
	if err := indexInverseOf(source, &index, predicates.InverseOf); err != nil {
		return nil, err
	}

	// and then read all the other data
	if err := indexData(source, &index); err != nil {
		return nil, err
	}

	// and finalize the index
	index.Finalize()
	return &index, nil
}

// indexSameAs inserts SameAs pairs into the index
func indexSameAs(source Source, index *Index, sameAsPredicates []string) error {
	if len(sameAsPredicates) == 0 {
		return nil
	}

	err := source.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	sameAss := make(map[string]struct{})
	for _, sameAs := range sameAsPredicates {
		sameAss[sameAs] = struct{}{}
	}

	for {
		tok := source.Next()
		switch {
		case tok.Err == io.EOF:
			return nil
		case tok.Err != nil:
			return tok.Err
		case !tok.HasDatum:
			if _, ok := sameAss[tok.Predicate]; ok {
				index.MarkIdentical(tok.Subject, tok.Object)
			}
		}
	}
}

// indexInverseOf inserts InverseOf pairs into the index
func indexInverseOf(source Source, index *Index, inversePredicates []string) error {
	if len(inversePredicates) == 0 {
		return nil
	}

	err := source.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	inverses := make(map[string]struct{})
	for _, inverse := range inversePredicates {
		inverses[inverse] = struct{}{}
	}

	for {
		tok := source.Next()
		switch {
		case tok.Err == io.EOF:
			return nil
		case tok.Err != nil:
			return tok.Err
		case !tok.HasDatum:
			if _, ok := inverses[tok.Predicate]; ok {
				index.MarkInverse(tok.Subject, tok.Object)
			}
		}
	}
}

// indexData inserts data into the index
func indexData(source Source, index *Index) error {
	err := source.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	for {
		tok := source.Next()
		switch {
		case tok.Err == io.EOF:
			return nil
		case tok.Err != nil:
			return tok.Err
		case tok.HasDatum:
			index.AddData(tok.Subject, tok.Predicate, tok.Datum)
		case !tok.HasDatum:
			index.AddTriple(tok.Subject, tok.Predicate, tok.Object)
		}
	}
}
