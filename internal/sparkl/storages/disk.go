package storages

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/iterator"
)

type DiskEngine struct {
	Path string
}

func (de DiskEngine) NewStorage(bundle *pathbuilder.Bundle) (BundleStorage, error) {
	path := filepath.Join(de.Path, bundle.Path.Bundle)

	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	return &Disk{
		DB: db,

		childStorages: make(map[string]BundleStorage, len(bundle.ChildBundles)),
	}, nil
}

// Disk represents a disk-backed storage
type Disk struct {
	count int64

	DB *leveldb.DB

	childStorages map[string]BundleStorage

	// l protects modifying data on disk
	l sync.RWMutex
}

func (ds *Disk) put(f func(*sEntity) error) error {
	entity := sEntityPool.Get().(*sEntity)
	entity.Reset()
	defer sEntityPool.Put(entity)

	if err := f(entity); err != nil {
		return err
	}

	data, err := entity.Encode()
	if err != nil {
		return err
	}

	ds.l.Lock()
	defer ds.l.Unlock()

	return ds.DB.Put([]byte(entity.URI), data, nil)
}

func (ds *Disk) get(uri wisski.URI, f func(*sEntity) error) error {
	entity := sEntityPool.Get().(*sEntity)
	entity.Reset()
	defer sEntityPool.Put(entity)

	ds.l.RLock()
	defer ds.l.RUnlock()

	// get the entity or return an error
	data, err := ds.DB.Get([]byte(uri), nil)
	if err == errors.ErrNotFound {
		return ErrNoEntity
	}
	if err != nil {
		return err
	}

	// decode the entity!
	if err := entity.Decode(data); err != nil {
		return err
	}

	// handle the entity!
	return f(entity)
}

func (ds *Disk) decode(data []byte, f func(*sEntity) error) error {
	entity := sEntityPool.Get().(*sEntity)
	entity.Reset()
	defer sEntityPool.Put(entity)

	if err := entity.Decode(data); err != nil {
		return err
	}

	return f(entity)
}

func (ds *Disk) update(uri wisski.URI, update func(*sEntity) error) error {
	entity := sEntityPool.Get().(*sEntity)
	entity.Reset()
	defer sEntityPool.Put(entity)

	ds.l.Lock()
	defer ds.l.Unlock()

	// get the entity or return an error
	data, err := ds.DB.Get([]byte(uri), nil)
	if err == errors.ErrNotFound {
		return ErrNoEntity
	}
	if err != nil {
		return err
	}

	// decode the entity!
	if err := entity.Decode(data); err != nil {
		return err
	}

	// perform the entity
	if err := update(entity); err != nil {
		return err
	}

	// encoded the entity again
	data, err = entity.Encode()
	if err != nil {
		return err
	}

	// and put it back!
	return ds.DB.Put([]byte(entity.URI), data, nil)
}

// Add adds an entity to this BundleSlice
func (ds *Disk) Add(uri wisski.URI, path []wisski.URI) error {
	atomic.AddInt64(&ds.count, 1)
	return ds.put(func(se *sEntity) error {
		se.URI = uri
		se.Path = path
		se.Fields = make(map[string][]wisski.FieldValue)
		se.Children = make(map[string][]wisski.URI)
		return nil
	})
}

func (ds *Disk) AddFieldValue(uri wisski.URI, field string, value any, path []wisski.URI) error {
	return ds.update(uri, func(se *sEntity) error {
		if se.Fields == nil {
			se.Fields = make(map[string][]wisski.FieldValue)
		}
		se.Fields[field] = append(se.Fields[field], wisski.FieldValue{
			Value: value,
			Path:  path,
		})
		return nil
	})
}

func (ds *Disk) RegisterChildStorage(bundle string, storage BundleStorage) error {
	ds.childStorages[bundle] = storage
	return nil
}

func (ds *Disk) AddChild(parent wisski.URI, bundle string, child wisski.URI) error {
	return ds.update(parent, func(se *sEntity) error {
		if se.Children == nil {
			se.Children = make(map[string][]wisski.URI)
		}
		se.Children[bundle] = append(se.Children[bundle], child)
		return nil
	})
}

func (ds *Disk) Finalize() error {
	return ds.DB.SetReadOnly()
}

func (ds *Disk) Get(parentPathIndex int) iterator.Iterator[URIWithParent] {
	return iterator.New(func(sender iterator.Generator[URIWithParent]) {
		defer sender.Return()

		it := ds.DB.NewIterator(nil, nil)
		defer it.Release()

		for it.Next() {
			var uri URIWithParent
			var err error

			if parentPathIndex > 0 {
				err = ds.decode(it.Value(), func(se *sEntity) error {
					uri.URI = se.URI
					if parentPathIndex != -1 {
						uri.Parent = se.Path[parentPathIndex]
					}
					return nil
				})
			} else {
				uri.URI = wisski.URI(it.Key())
			}

			if sender.YieldError(err) {
				return
			}

			if sender.Yield(uri) {
				return
			}
		}

		sender.YieldError(it.Error())
	})
}

func (ds *Disk) Count() (int64, error) {
	return atomic.LoadInt64(&ds.count), nil
}

func (ds *Disk) Load(uri wisski.URI) (entity wisski.Entity, err error) {
	err = ds.get(uri, func(se *sEntity) error {
		// copy simple fields
		entity.URI = se.URI
		entity.Path = se.Path
		entity.Fields = se.Fields

		// load all the child entities
		entity.Children = make(map[string][]wisski.Entity)
		for bundle, value := range se.Children {
			entity.Children[bundle] = make([]wisski.Entity, len(value))
			for i, uri := range value {
				entity.Children[bundle][i], err = ds.childStorages[bundle].Load(uri)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return
}

func (ds *Disk) Close() error {
	if ds.DB != nil {
		ds.DB.Close()
		ds.DB = nil
		ds.childStorages = nil
	}
	return nil
}