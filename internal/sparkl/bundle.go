package sparkl

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// ExtractEntities loads all entities from the given bundle into a new storage, which is then returned.
//
// Storages for any child bundles, and the bundle itself, are created using the makeStorage function.
// The storage for this bundle is returned.
func ExtractEntities(bundle *pathbuilder.Bundle, index *Index, makeStorage func(bundle *pathbuilder.Bundle) BundleStorage) BundleStorage {
	// initialize a new storage for this bundle
	storage := makeStorage(bundle)

	// determine the index of the URI within the paths describing this bundle
	// this is the length of the parent path, or zero (if it does not exist).
	var entityURIIndex int
	if bundle.Parent != nil {
		entityURIIndex = len(bundle.Group.PathArray) / 2
	}

	var wg sync.WaitGroup

	// add all the elements
	// adding element is performed concurrently.
	for path := range extractPath(bundle.Group, index) {
		nodes := path.Nodes()
		storage.Add(nodes[entityURIIndex], nodes)
	}

	// scan all the paths for all of the fields
	for _, field := range bundle.Fields() {
		wg.Add(1)
		go func(field pathbuilder.Field) {
			defer wg.Done()

			for path := range extractPath(field.Path, index) {
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

	// fetch all the child bundles
	for _, bundle := range bundle.ChildBundles {
		wg.Add(1)
		go func(bundle *pathbuilder.Bundle) {
			defer wg.Done()

			storage := ExtractEntities(bundle, index, makeStorage)
			defer storage.Close()

			for entity := range storage.Get() {
				uri := entity.Path[entityURIIndex]
				storage.AddChild(uri, bundle.Group.ID, entity)
			}
		}(bundle)
	}

	wg.Wait()

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
