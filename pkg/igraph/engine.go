package igraph

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
)

// Engine represents an object that creates storages for an IGraph
type Engine[Label comparable, Datum any] interface {
	imap.Engine[Label]

	Data() (imap.Storage[imap.ID, Datum], error)
	Inverses() (imap.Storage[imap.ID, imap.ID], error)
	PSOIndex() (ThreeStorage, error)
	POSIndex() (ThreeStorage, error)
}

type ThreeStorage interface {
	io.Closer

	// Add adds a new mapping for the given ids
	Add(a, b, c imap.ID) error

	// Count counts the overall number of entries in the index
	Count() (int64, error)

	// Finalize informs the storage that no more mappings will be made
	Finalize() error

	// Fetch iterates over all triples (a, b, c) in c-order.
	// If an error occurs, iteration stops and is returned to the caller
	Fetch(a, b imap.ID, f func(c imap.ID) error) error

	// Has checks if the given mapping exists
	Has(a, b, c imap.ID) (bool, error)
}
