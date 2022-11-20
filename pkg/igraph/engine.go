package igraph

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
)

// Engine represents an object that creates storages for an IGraph
type Engine[Label comparable, Datum any] interface {
	imap.Engine[Label]

	Data() (imap.KeyValueStore[imap.ID, Datum], error)
	Triples() (imap.KeyValueStore[imap.ID, IndexTriple], error)
	Inverses() (imap.KeyValueStore[imap.ID, imap.ID], error)
	PSOIndex() (ThreeStorage, error)
	POSIndex() (ThreeStorage, error)
}

type ThreeStorage interface {
	io.Closer

	// Add adds a new mapping for the given (a, b, c).
	//
	// l acts as a label for the insert.
	// when the given edge already exists, the conflict function should be called to resolve the conflict
	Add(a, b, c imap.ID, l imap.ID, conflict func(old, new imap.ID) (imap.ID, error)) (conflicted bool, err error)

	// Count counts the overall number of entries in the index
	Count() (int64, error)

	// Finalize informs the storage that no more mappings will be made
	Finalize() error

	// Fetch iterates over all triples (a, b, c) in c-order.
	// l is the last label that was created for the triple.
	// If an error occurs, iteration stops and is returned to the caller
	Fetch(a, b imap.ID, f func(c imap.ID, l imap.ID) error) error

	// Has checks if the given mapping exists
	Has(a, b, c imap.ID) (bool, error)
}
