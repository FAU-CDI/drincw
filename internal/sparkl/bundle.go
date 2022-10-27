package sparkl

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// StoreBundle loads all entities from the given bundle into a new storage, which is then returned.
//
// Storages for any child bundles, and the bundle itself, are created using the makeStorage function.
// The storage for this bundle is returned.
func StoreBundle(bundle *pathbuilder.Bundle, index *Index, makeStorage func(bundle *pathbuilder.Bundle) BundleStorage) BundleStorage {
	return StoreBundles([]*pathbuilder.Bundle{bundle}, index, makeStorage)[0]
}

// StoreBundles is like StoreBundle, but takes multiple bundles
func StoreBundles(bundles []*pathbuilder.Bundle, index *Index, makeStorage func(bundle *pathbuilder.Bundle) BundleStorage) []BundleStorage {
	var extractWait, childAddWait sync.WaitGroup
	context := &context{
		index:       index,
		makeStorage: makeStorage,

		extractWait:  &extractWait,
		childAddWait: &childAddWait,
	}

	storages := make([]BundleStorage, len(bundles))
	for i := range storages {
		extractWait.Add(1)
		go func(i int) {
			storages[i] = storeBundle(bundles[i], context)
		}(i)
	}

	extractWait.Wait()
	childAddWait.Wait()
	return storages
}

type context struct {
	index       *Index
	makeStorage func(bundle *pathbuilder.Bundle) BundleStorage

	extractWait  *sync.WaitGroup // waiting on extracting entities in all bundles
	childAddWait *sync.WaitGroup // loading child entities wait
}

func storeBundle(bundle *pathbuilder.Bundle, context *context) BundleStorage {
	defer context.extractWait.Done()

	// initialize a new storage for this bundle
	storage := context.makeStorage(bundle)

	// determine the index of the URI within the paths describing this bundle
	// this is the length of the parent path, or zero (if it does not exist).
	var entityURIIndex int
	if bundle.Parent != nil {
		entityURIIndex = len(bundle.Group.PathArray) / 2
	}

	// stage 1: load the entities themselves
	for path := range extractPath(bundle.Group, context.index) {
		nodes := path.Nodes()
		storage.Add(nodes[entityURIIndex], nodes)
	}

	// stage 2: fill all the fields
	for _, field := range bundle.Fields() {
		context.extractWait.Add(1)
		go func(field pathbuilder.Field) {
			defer context.extractWait.Done()

			for path := range extractPath(field.Path, context.index) {
				nodes := path.Nodes()
				datum, hasDatum := path.Datum()
				if !hasDatum && len(nodes) > 0 {
					datum = nodes[len(nodes)-1]
				}
				uri := nodes[entityURIIndex]

				storage.AddFieldValue(uri, field.ID, datum, nodes)
			}
		}(field)
	}

	// stage 3: read child paths
	storages := make([]BundleStorage, len(bundle.ChildBundles))
	for i, bundle := range bundle.ChildBundles {
		context.extractWait.Add(1)
		context.childAddWait.Add(1)

		go func(i int, bundle *pathbuilder.Bundle) {
			storages[i] = storeBundle(bundle, context) // does context.extractWait.Done()
		}(i, bundle)
	}

	// stage 4: register all the child entities
	go func() {
		context.extractWait.Wait()

		for i, cstorage := range storages {
			go func(cstorage BundleStorage, bundle *pathbuilder.Bundle) {
				defer context.childAddWait.Done()

				for child := range cstorage.Get(entityURIIndex) {
					storage.AddChild(child.Parent, bundle.Group.ID, child.URI, cstorage)
				}
				cstorage.Close()
			}(cstorage, bundle.ChildBundles[i])
		}
	}()

	return storage
}

const (
	debugLogAllPaths = false   // turn this on to log all paths being queried
	datatypeEmpty    = "empty" // a datatype being recalled as "empty"
)

var debugLogID int64 // id of the current log id

// extractPath extracts values for a single path from the index
func extractPath(path pathbuilder.Path, index *Index) <-chan Path {
	// start with the path array
	uris := append([]string{}, path.PathArray...)
	if len(uris) == 0 {
		return nil
	}

	// add the datatype property if are not a group
	// and it is not empty
	if !path.IsGroup && path.DatatypeProperty != "" && path.DatatypeProperty != datatypeEmpty {
		uris = append(uris, path.DatatypeProperty)
	}

	// if debugging is enabled, set it up
	var debugID int64
	if debugLogAllPaths {
		debugID = atomic.AddInt64(&debugLogID, 1)
	}

	set := index.PathsStarting(Type, URI(uris[0]))
	if debugLogAllPaths {
		log.Println(debugID, uris[0], set.Size())
	}

	for i := 1; i < len(uris) && set.Size() > 0; i++ {
		if i%2 == 0 {
			set.Ending(Type, URI(uris[i]))
		} else {
			set.Connected(URI(uris[i]))
		}

		if debugLogAllPaths {
			log.Println(debugID, uris[i], set.Size())
		}
	}

	return set.Paths()
}
