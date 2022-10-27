package sparkl

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// BundleStorage is responsible for storing entities for a single bundle
type BundleStorage interface {
	// Add adds a new entity with the given URI (and optional path information)
	// to this bundle.
	//
	// Calls to add for a specific bundle storage are serialized.
	Add(uri URI, path []URI)

	// AddFieldValue adds a value to the given field for the entity with the given uri.
	AddFieldValue(uri URI, field string, value any, path []URI)

	// AddChild adds a child entity of the given bundle to the given entity.
	//
	// Multiple concurrent calls to AddFieldValue may take place.
	AddChild(uri URI, bundle string, entity Entity)

	// Get returns a channel that receives each entity that was created.
	// Once all entities have been returned, the channel is closed.
	Get() <-chan Entity

	// Load loads an entity with the given URI from this storage
	Load(uri URI) Entity

	// Close closes this BundleStorage
	Close()
}

// NewBundleSlice create a new BundleSlice storage.
// It can be used as the makeStorage argument to [ExtractBundles].
func NewBundleSlice(bundle *pathbuilder.Bundle) BundleStorage {
	return &BundleSlice{
		bundle: bundle,
		lookup: make(map[URI]int),
		adding: make(chan struct{}),
	}
}

// BundleSlice implements an in-memory BundleStorage
type BundleSlice struct {
	Entities []Entity

	bundle *pathbuilder.Bundle

	setFieldLock sync.Mutex
	addChildLock sync.Mutex

	lookup map[URI]int
	adding chan struct{}
}

// Add adds an entity to this BundleSlice
func (bs *BundleSlice) Add(uri URI, path []URI) {
	bs.lookup[uri] = len(bs.Entities)
	entity := Entity{
		URI:      uri,
		Path:     path,
		Fields:   make(map[string][]FieldValue, len(bs.bundle.ChildFields)),
		Children: make(map[string][]Entity, len(bs.bundle.ChildBundles)),
	}

	for _, field := range bs.bundle.ChildFields {
		entity.Fields[field.ID] = make([]FieldValue, 0, field.MakeCardinality())
	}

	for _, bundle := range bs.bundle.ChildBundles {
		entity.Children[bundle.Group.ID] = make([]Entity, 0, bundle.Group.MakeCardinality())
	}

	bs.Entities = append(bs.Entities, entity)
}

// AddFieldValue
func (bs *BundleSlice) AddFieldValue(uri URI, field string, value any, path []URI) {
	bs.setFieldLock.Lock()
	defer bs.setFieldLock.Unlock()

	id, ok := bs.lookup[uri]
	if !ok {
		return
	}

	if bs.Entities[id].Fields == nil {
		bs.Entities[id].Fields = make(map[string][]FieldValue)
	}
	bs.Entities[id].Fields[field] = append(bs.Entities[id].Fields[field], FieldValue{
		Value: value,
		Path:  path,
	})
}

func (bs *BundleSlice) AddChild(uri URI, bundle string, entity Entity) {
	bs.addChildLock.Lock()
	defer bs.addChildLock.Unlock()

	id, ok := bs.lookup[uri]
	if !ok {
		return
	}

	if bs.Entities[id].Children == nil {
		bs.Entities[id].Children = make(map[string][]Entity)
	}
	bs.Entities[id].Children[bundle] = append(bs.Entities[id].Children[bundle], entity)
}

func (bs *BundleSlice) Get() <-chan Entity {
	c := make(chan Entity)
	go func() {
		defer close(c)
		for _, entity := range bs.Entities {
			c <- entity
		}
	}()
	return c
}

func (bs *BundleSlice) Load(uri URI) Entity {
	return bs.Entities[bs.lookup[uri]]
}

func (bs *BundleSlice) Close() {
	bs.lookup = nil
}
