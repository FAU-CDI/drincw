package storages

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// NewBundleSlice create a new BundleSlice storage.
// It can be used as the makeStorage argument to [ExtractBundles].
func NewBundleSlice(bundle *pathbuilder.Bundle) (BundleStorage, error) {
	return &BundleSlice{
		bundle: bundle,
		lookup: make(map[wisski.URI]int),
		adding: make(chan struct{}),
	}, nil
}

// BundleSlice implements an in-memory BundleStorage
type BundleSlice struct {
	Entities []wisski.Entity

	bundle *pathbuilder.Bundle

	setFieldLock sync.Mutex
	addChildLock sync.Mutex

	lookup map[wisski.URI]int
	adding chan struct{}
}

// Add adds an entity to this BundleSlice
func (bs *BundleSlice) Add(uri wisski.URI, path []wisski.URI) error {
	bs.lookup[uri] = len(bs.Entities)
	entity := wisski.Entity{
		URI:      uri,
		Path:     path,
		Fields:   make(map[string][]wisski.FieldValue, len(bs.bundle.ChildFields)),
		Children: make(map[string][]wisski.Entity, len(bs.bundle.ChildBundles)),
	}

	for _, field := range bs.bundle.ChildFields {
		entity.Fields[field.ID] = make([]wisski.FieldValue, 0, field.MakeCardinality())
	}

	for _, bundle := range bs.bundle.ChildBundles {
		entity.Children[bundle.Group.ID] = make([]wisski.Entity, 0, bundle.Group.MakeCardinality())
	}

	bs.Entities = append(bs.Entities, entity)
	return nil
}

// AddFieldValue
func (bs *BundleSlice) AddFieldValue(uri wisski.URI, field string, value any, path []wisski.URI) error {
	bs.setFieldLock.Lock()
	defer bs.setFieldLock.Unlock()

	id, ok := bs.lookup[uri]
	if !ok {
		return ErrNoEntity
	}

	bs.Entities[id].Fields[field] = append(bs.Entities[id].Fields[field], wisski.FieldValue{
		Value: value,
		Path:  path,
	})

	return nil
}

func (bs *BundleSlice) AddChild(parent wisski.URI, bundle string, child wisski.URI, storage BundleStorage) error {
	bs.addChildLock.Lock()
	defer bs.addChildLock.Unlock()

	id, ok := bs.lookup[parent]
	if !ok {
		return ErrNoEntity
	}

	entity, err := storage.Load(child)
	if err != nil {
		return err
	}
	bs.Entities[id].Children[bundle] = append(bs.Entities[id].Children[bundle], entity)
	return nil
}

func (bs *BundleSlice) Get(parentPathIndex int, errDst *error) <-chan URIWithParent {
	c := make(chan URIWithParent)
	go func() {
		defer close(c)
		for _, entity := range bs.Entities {
			var parent wisski.URI
			if parentPathIndex > -1 {
				parent = entity.Path[parentPathIndex]
			}
			c <- URIWithParent{
				URI:    entity.URI,
				Parent: parent,
			}
		}
	}()
	return c
}

func (bs *BundleSlice) Load(uri wisski.URI) (entity wisski.Entity, err error) {
	index, ok := bs.lookup[uri]
	if !ok {
		return entity, ErrNoEntity
	}
	return bs.Entities[index], nil
}

func (bs *BundleSlice) Close() error {
	bs.lookup = nil
	return nil
}