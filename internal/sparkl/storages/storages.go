package storages

import (
	"errors"
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// BundleEngine is a function that initializes and returns a new BundleStorage
type BundleEngine func(bundle *pathbuilder.Bundle) (BundleStorage, error)

// BundleStorage is responsible for storing entities for a single bundle
type BundleStorage interface {
	io.Closer

	// Add adds a new entity with the given URI (and optional path information)
	// to this bundle.
	//
	// Calls to add for a specific bundle storage are serialized.
	Add(uri wisski.URI, path []wisski.URI) error

	// AddFieldValue adds a value to the given field for the entity with the given uri.
	//
	// A non-existing uri should return ErrNoEntity.
	AddFieldValue(uri wisski.URI, field string, value any, path []wisski.URI) error

	// AddChild adds a child entity of the given bundle to the given entity.
	//
	// Multiple concurrent calls to AddChild may take place.
	//
	// A non-existing parent should return ErrNoEntity.
	AddChild(parent wisski.URI, bundle string, child wisski.URI, storage BundleStorage) error

	// Get returns a channel that receives the url of every entity in this bundle, along with their parent URIs.
	// parentPathIndex returns the index of the parent uri in child paths
	//
	// The caller is responsible for draining the channel.
	Get(parentPathIndex int, errDst *error) <-chan URIWithParent

	// Load loads an entity with the given URI from this storage.
	// A non-existing entity should return err = ErrNoEntity.
	Load(uri wisski.URI) (wisski.Entity, error)
}

var (
	ErrNoEntity = errors.New("No such entity")
)

// URIWithParent represents a URI along with it's parent
type URIWithParent struct {
	URI    wisski.URI
	Parent wisski.URI
}
