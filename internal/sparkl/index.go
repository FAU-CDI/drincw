// Package sparkl provides facilities to generate an Index for a WissKI
package sparkl

import (
	"io"
	"os"
)

// LoadIndex is like MakeIndex, but reads nquads from the given path.
// When err != nil, the caller must eventually close the index.
func LoadIndex(path string, predicates Predicates, engine Engine) (*Index, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return MakeIndex(&QuadSource{Reader: reader}, predicates, engine)
}

// MakeIndex creates a new Index from the given source.
// When err != nil, the caller must eventually close the index.
func MakeIndex(source Source, predicates Predicates, engine Engine) (*Index, error) {
	// create a new index
	var index Index
	if err := index.Reset(engine); err != nil {
		return nil, err
	}

	// read the "same as" triples first
	if err := indexSameAs(source, &index, predicates.SameAs); err != nil {
		index.Close()
		return nil, err
	}

	// read the "inverse" triples next
	if err := indexInverseOf(source, &index, predicates.InverseOf); err != nil {
		index.Close()
		return nil, err
	}

	// and then read all the other data
	if err := indexData(source, &index); err != nil {
		index.Close()
		return nil, err
	}

	// and finalize the index
	if err := index.Finalize(); err != nil {
		index.Close()
		return nil, err
	}
	return &index, nil
}

// indexSameAs inserts SameAs pairs into the index
func indexSameAs(source Source, index *Index, sameAsPredicates []URI) error {
	if len(sameAsPredicates) == 0 {
		return nil
	}

	err := source.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	sameAss := make(map[URI]struct{})
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
func indexInverseOf(source Source, index *Index, inversePredicates []URI) error {
	if len(inversePredicates) == 0 {
		return nil
	}

	err := source.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	inverses := make(map[URI]struct{})
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
