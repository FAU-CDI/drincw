package sparkl

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// BundleStorage is responsible for storing entities for a single bundle
type BundleStorage interface {
	// Add adds a new Entity to this storage.
	// Only the URI and Path fields will be filled.
	//
	// Calls to Add will be serialized.
	Add(Entity)

	// DoneAdding signals that no more entities will be added to this BundleStorage.
	DoneAdding()

	// AddFieldValue adds a value to the given field for the entity with the given uri.
	AddFieldValue(uri URI, field string, value FieldValue)

	// AddChild adds a child entity of the given bundle to the given entity.
	//
	// Multiple concurrent calls to AddFieldValue may take place.
	AddChild(uri URI, bundle string, entity Entity)

	// Done informs this BundleStorage that no more calls will be made to Add, AddFieldValue, AddChild
	DoneStoring()

	// Get returns a channel that receives each entity that was created.
	// Once all entities have been returned, the channel is closed.
	Get() <-chan Entity
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
func (bs *BundleSlice) Add(entity Entity) {
	uri := entity.URI
	bs.lookup[uri] = len(bs.Entities)
	bs.Entities = append(bs.Entities, entity)
}

// DoneAdding signals that no more adds should take place
func (bs *BundleSlice) DoneAdding() {
	close(bs.adding)
}

// AddFieldValue
func (bs *BundleSlice) AddFieldValue(uri URI, field string, value FieldValue) {
	<-bs.adding
	bs.setFieldLock.Lock()
	defer bs.setFieldLock.Unlock()

	id, ok := bs.lookup[uri]
	if !ok {
		return
	}

	if bs.Entities[id].Fields == nil {
		bs.Entities[id].Fields = make(map[string][]FieldValue)
	}
	bs.Entities[id].Fields[field] = append(bs.Entities[id].Fields[field], value)
}

func (bs *BundleSlice) AddChild(uri URI, bundle string, entity Entity) {
	<-bs.adding
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

func (bs *BundleSlice) DoneStoring() {
	bs.lookup = nil
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
