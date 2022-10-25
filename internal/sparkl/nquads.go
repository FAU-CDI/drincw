package sparkl

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type Index = igraph.IGraph[string, any]

const SameAs = "http://www.w3.org/2002/07/owl#sameAs"

// ReadNQuads
func ReadNQuads(r io.ReadSeeker, SameAsPredicates []string) (*Index, error) {
	// create a new index
	var index Index

	source := QuadReadSeeker{Reader: r}
	index.Reset()

	// read the "same as" triples first
	if err := readSameAs(&source, &index, SameAsPredicates); err != nil {
		return nil, err
	}

	index.ApplyIdentifications()

	// and then read all the other data
	if err := readData(&source, &index); err != nil {
		return nil, err
	}

	// and finalize the index
	index.Finalize()
	return &index, nil
}

func readSameAs(source Source, index *Index, sameAsPredicates []string) error {
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
				index.Identify(tok.Subject, tok.Object)
			}
		}
	}
}

func readData(source Source, index *Index) error {
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
