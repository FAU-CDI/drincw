package igraph

import (
	"path/filepath"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
)

// DiskEngine represents an engine that stores everything on disk
type DiskEngine[Label comparable, Datum any] struct {
	imap.DiskEngine[Label]

	MarshalDatum   func(datum Datum) ([]byte, error)
	UnmarshalDatum func(dest *Datum, src []byte) error
}

func (de DiskEngine[Label, Datum]) Data() (imap.Storage[imap.ID, Datum], error) {
	data := filepath.Join(de.Path, "igraph_data.pogrep")

	ds, err := imap.NewDiskStorage[imap.ID, Datum](data, de.Options)
	if err != nil {
		return nil, err
	}

	ds.MarshalKey = imap.MarshalID
	ds.UnmarshalKey = imap.UnmarshalID

	if de.MarshalDatum != nil && de.UnmarshalDatum != nil {
		ds.MarshalValue = de.MarshalDatum
		ds.UnmarshalValue = de.UnmarshalDatum
	}

	return ds, nil
}
func (de DiskEngine[Label, Datum]) Inverses() (imap.Storage[imap.ID, imap.ID], error) {
	inverses := filepath.Join(de.Path, "igraph_inverses.pogrep")

	ds, err := imap.NewDiskStorage[imap.ID, imap.ID](inverses, de.Options)
	if err != nil {
		return nil, err
	}

	ds.MarshalKey = imap.MarshalID
	ds.UnmarshalKey = imap.UnmarshalID

	ds.MarshalValue = imap.MarshalID
	ds.UnmarshalValue = imap.UnmarshalID

	return ds, nil
}
func (DiskEngine[Label, Datum]) PSOIndex() (ThreeStorage, error) {
	th := make(ThreeHash)
	return &th, nil // TODO: Make this actually on disk

}
func (DiskEngine[Label, Datum]) POSIndex() (ThreeStorage, error) {
	th := make(ThreeHash)
	return &th, nil // TODO: Make this actually on disk
}
