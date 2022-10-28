package igraph

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Engine represents an object that creates storages for an IGraph
type Engine[Label comparable, Datum any] interface {
	imap.Engine[Label]

	Data() (imap.Storage[imap.ID, Datum], error)
	Inverses() (imap.Storage[imap.ID, imap.ID], error)
	PSOIndex() (ThreeStorage, error)
	POSIndex() (ThreeStorage, error)
}

// MemoryEngine represents an engine that stores everything in memory
type MemoryEngine[Label comparable, Datum any] struct {
	imap.MemoryEngine[Label]
}

func (MemoryEngine[Label, Datum]) Data() (imap.Storage[imap.ID, Datum], error) {
	return make(imap.MemoryStorage[imap.ID, Datum]), nil
}
func (MemoryEngine[Label, Datum]) Inverses() (imap.Storage[imap.ID, imap.ID], error) {
	return make(imap.MemoryStorage[imap.ID, imap.ID]), nil
}
func (MemoryEngine[Label, Datum]) PSOIndex() (ThreeStorage, error) {
	return make(ThreeHash), nil

}
func (MemoryEngine[Label, Datum]) POSIndex() (ThreeStorage, error) {
	return make(ThreeHash), nil
}

type ThreeStorage interface {
	io.Closer

	// Add adds a new mapping for the given ids
	Add(a, b, c imap.ID) error

	// Count counts the overall number of entries in the index
	Count() (int64, error)

	// Finalize informs the storage that no more mappings will be made
	Finalize() error

	// Fetch iterates over all triples (a, b, c) in insertion order on c.
	// If an error occurs, iteration stops and is returned to the caller
	Fetch(a, b imap.ID, f func(c imap.ID) error) error

	// Has checks if the given mapping exists
	Has(a, b, c imap.ID) (bool, error)
}

// ThreeHash implements ThreeStorage in memory
type ThreeHash map[imap.ID]map[imap.ID]*ThreeItem

type ThreeItem struct {
	Keys []imap.ID
	Data map[imap.ID]struct{}
}

func (tlm ThreeHash) Add(a, b, c imap.ID) error {
	switch {
	case tlm[a] == nil:
		tlm[a] = make(map[imap.ID]*ThreeItem)
		fallthrough
	case tlm[a][b] == nil:
		tlm[a][b] = &ThreeItem{
			Data: make(map[imap.ID]struct{}, 1),
		}
		fallthrough
	default:
		tlm[a][b].Data[c] = struct{}{}
	}
	return nil
}

func (tlm ThreeHash) Count() (total int64, err error) {
	for _, a := range tlm {
		for _, b := range a {
			total += int64(len(b.Keys))
		}
	}
	return total, nil
}

func (tlm ThreeHash) Finalize() error {
	for _, a := range tlm {
		for _, b := range a {
			b.Keys = maps.Keys(b.Data)
			slices.SortFunc(b.Keys, func(a imap.ID, b imap.ID) bool {
				return a.Less(b)
			})
		}
	}
	return nil
}

func (tlm ThreeHash) Fetch(a, b imap.ID, f func(c imap.ID) error) error {
	three := tlm[a][b]
	if three == nil {
		return nil
	}
	for _, c := range three.Keys {
		if err := f(c); err != nil {
			return err
		}
	}
	return nil
}

func (tlm ThreeHash) Has(a, b, c imap.ID) (bool, error) {
	three := tlm[a][b]
	if three == nil {
		return false, nil
	}
	_, ok := three.Data[c]
	return ok, nil
}

func (tlm ThreeHash) Close() error {
	return nil
}
