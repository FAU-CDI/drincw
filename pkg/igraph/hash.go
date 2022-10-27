package igraph

import (
	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type ThreeStorage interface {
	// Add adds a new mapping for the given ids
	Add(a, b, c imap.ID)

	// Count counts the overall number of entries in the index
	Count() int64

	// Finalize informs the storage that no more mappings will be made
	Finalize()

	// Fetch iterates over all triples (a, b, c) in insertion order on c.
	Fetch(a, b imap.ID, f func(c imap.ID))

	// Has checks if the given mapping exists
	Has(a, b, c imap.ID) bool // posIndex
}

// ThreeHash implements ThreeStorage in memory
type ThreeHash map[imap.ID]map[imap.ID]*ThreeItem

type ThreeItem struct {
	Keys []imap.ID
	Data map[imap.ID]struct{}
}

func (tlm ThreeHash) Add(a, b, c imap.ID) {
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
}

func (tlm ThreeHash) Count() (total int64) {
	for _, a := range tlm {
		for _, b := range a {
			total += int64(len(b.Keys))
		}
	}
	return total
}

func (tlm ThreeHash) Finalize() {
	for _, a := range tlm {
		for _, b := range a {
			b.Keys = maps.Keys(b.Data)
			slices.SortFunc(b.Keys, func(a imap.ID, b imap.ID) bool {
				return a.Less(b)
			})
		}
	}
}

func (tlm ThreeHash) Fetch(a, b imap.ID, f func(c imap.ID)) {
	three := tlm[a][b]
	if three == nil {
		return
	}
	for _, c := range three.Keys {
		f(c)
	}
}

func (tlm ThreeHash) Has(a, b, c imap.ID) bool {
	three := tlm[a][b]
	if three == nil {
		return false
	}
	_, ok := three.Data[c]
	return ok
}