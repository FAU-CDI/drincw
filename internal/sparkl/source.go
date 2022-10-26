package sparkl

import (
	"io"

	"github.com/cayleygraph/quad"
	"github.com/cayleygraph/quad/nquads"
)

// Source represents a source of triple data
type Source interface {
	// Open opens this data source.
	//
	// It is valid to call open more than once after Next() returns a token with err = io.EOF.
	// In this case the second call to open should reset the data source.
	Open() error

	// Close closes this source.
	// Close may only be called once a token with err != io.EOF is called.
	Close() error

	// Next scans the next token
	Next() Token
}

// Token represents a token read from a triplestore file.
//
// It can represent one of three states:
//
// 1. an error token
// 1. a (subject, predicate, object) token
// 2. a (subject, predicate, datum) token
//
// In the case of 1, Error != nil.
// In the case of 2, Error == nil && HasDatum = False
// In the case of 3, Error == nil && HasDatum = True
type Token struct {
	Subject   string
	Predicate string
	Object    string

	HasDatum bool
	Datum    any

	Err error
}

// QuadSource reads triples from a quad file
type QuadSource struct {
	Reader io.ReadSeeker
	reader *nquads.Reader
}

func (qs *QuadSource) Open() error {
	// if we previously had a reader
	// then we need to reset the state
	if qs.reader != nil {
		if err := qs.reader.Close(); err != nil {
			return err
		}
		_, err := qs.Reader.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	}

	qs.reader = nquads.NewReader(qs.Reader, true)
	return nil
}

// Next reads the next token from the QuadSource
func (qs *QuadSource) Next() Token {
	for {
		value, err := qs.reader.ReadQuad()
		if err != nil {
			return Token{Err: err}
		}

		sI, sOK := asIRILike(value.Subject.Native())
		pI, pOK := asIRILike(value.Predicate.Native())
		if !(sOK && pOK) {
			continue
		}

		o := value.Object.Native()
		oI, oOK := asIRILike(o)
		if oOK {
			return Token{
				Subject:   sI,
				Predicate: pI,
				Object:    oI,
			}
		} else {
			return Token{
				Subject:   sI,
				Predicate: pI,
				HasDatum:  true,
				Datum:     o,
			}
		}
	}
}

func (qs *QuadSource) Close() error {
	if qs.reader != nil {
		return qs.reader.Close()
	}
	return nil
}

func asIRILike(value any) (uri string, ok bool) {
	switch datum := value.(type) {
	case quad.IRI:
		return string(datum), true
	case quad.BNode:
		return string(datum), true
	default:
		return "", false
	}
}
