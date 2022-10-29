package igraph

import (
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
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
func (de DiskEngine[Label, Datum]) PSOIndex() (ThreeStorage, error) {
	pso := filepath.Join(de.Path, "pso.leveldb")
	return NewDiskHash(pso)
}
func (de DiskEngine[Label, Datum]) POSIndex() (ThreeStorage, error) {
	pos := filepath.Join(de.Path, "pos.leveldb")
	return NewDiskHash(pos)
}

func NewDiskHash(path string) (ThreeStorage, error) {
	// If the path already exists, wipe it
	_, err := os.Stat(path)
	if err == nil {
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	level, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, err
	}
	dh := &ThreeDiskHash{
		DB: level,
	}
	return dh, nil
}

// ThreeHash implements ThreeStorage in memory
type ThreeDiskHash struct {
	DB   *leveldb.DB
	ropt opt.ReadOptions
	wopt opt.WriteOptions
}

// encodeTriple encodes the given id triple into a range
func encodeTriple(id1, id2, id3 imap.ID) []byte {
	b := make([]byte, 24)
	binary.BigEndian.PutUint64(b[0:8], id1[0])
	binary.BigEndian.PutUint64(b[8:16], id2[0])
	binary.BigEndian.PutUint64(b[16:24], id3[0])
	return b
}

func decodeLast(data []byte) imap.ID {
	third := binary.BigEndian.Uint64(data[16:24])
	return [1]uint64{third}
}

var (
	minID = [1]uint64{0}
	maxID = [1]uint64{^uint64(0)}
)

func (tlm *ThreeDiskHash) Add(a, b, c imap.ID) error {
	return tlm.DB.Put(encodeTriple(a, b, c), nil, &tlm.wopt)
}

func (tlm *ThreeDiskHash) Count() (total int64, err error) {
	iterator := tlm.DB.NewIterator(nil, &tlm.ropt)
	defer iterator.Release()

	for iterator.Next() {
		total++
	}

	if err := iterator.Error(); err != nil {
		return 0, err
	}

	return total, nil
}

func (tlm ThreeDiskHash) Finalize() error {
	if err := tlm.DB.CompactRange(util.Range{}); err != nil {
		return err
	}
	return tlm.DB.SetReadOnly()
}

func (tlm *ThreeDiskHash) Fetch(a, b imap.ID, f func(c imap.ID) error) error {
	iterator := tlm.DB.NewIterator(&util.Range{
		Start: encodeTriple(a, b, minID),
		Limit: encodeTriple(a, b, maxID),
	}, &tlm.ropt)
	defer iterator.Release()

	for iterator.Next() {
		c := decodeLast(iterator.Key())
		if err := f(c); err != nil {
			return err
		}
	}

	if err := iterator.Error(); err != nil {
		return err
	}

	return nil
}

func (tlm *ThreeDiskHash) Has(a, b, c imap.ID) (bool, error) {
	return tlm.DB.Has(encodeTriple(a, b, c), &tlm.ropt)
}

func (tlm *ThreeDiskHash) Close() (err error) {
	if tlm.DB != nil {
		err = tlm.DB.Close()
		tlm.DB = nil
	}
	return
}
