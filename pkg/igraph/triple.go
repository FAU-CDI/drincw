package igraph

import (
	"errors"
	"fmt"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
)

// Stats holds statistics about triples in the index
type Stats struct {
	DirectTriples   int64
	DatumTriples    int64
	InverseTriples  int64
	ConflictTriples int64
}

func (stats Stats) String() string {
	return fmt.Sprintf("{direct:%d,datum:%d,inverse:%d,conflict:%d}", stats.DirectTriples, stats.DatumTriples, stats.InverseTriples, stats.ConflictTriples)
}

// IndexTriple represents a triple stored inside the index
type IndexTriple struct {
	Role             // Why was this triple stored?
	Items [3]imap.ID // What were the original items in this triple?
}

func MarshalTriple(triple IndexTriple) ([]byte, error) {
	result := make([]byte, 3*imap.IDLen+1)
	triple.Items[0].Encode(result[:imap.IDLen])
	triple.Items[1].Encode(result[imap.IDLen : 2*imap.IDLen])
	triple.Items[2].Encode(result[2*imap.IDLen:])
	result[len(result)-1] = byte(triple.Role)
	return result, nil
}

var errDecodeTriple = errors.New("DecodeTriple: src too short")

func UnmarshalTriple(dest *IndexTriple, src []byte) error {
	if len(src) < 3*imap.IDLen+1 {
		return errDecodeTriple
	}
	dest.Items[0].Decode(src[:imap.IDLen])
	dest.Items[1].Decode(src[imap.IDLen : 2*imap.IDLen])
	dest.Items[2].Decode(src[2*imap.IDLen:])
	dest.Role = Role(src[3*imap.IDLen])
	return nil
}

// Triple represents a resolve triple
type Triple[Label comparable, Datum any] struct {
	Role Role

	Subject, Predicate, Object Label

	Datum Datum
}

// Role represents the role of the triple
type Role uint8

const (
	// Regular represents a regular (non-infered) triple
	Regular Role = iota

	// Inverse represents an infered inverse triple
	Inverse

	// Data represents a data triple
	Data
)
