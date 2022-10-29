package sparkl

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// StoreBundle loads all entities from the given bundle into a new storage, which is then returned.
//
// Storages for any child bundles, and the bundle itself, are created using the makeStorage function.
// The storage for this bundle is returned.
func StoreBundle(bundle *pathbuilder.Bundle, index *Index, engine BundleEngine) (BundleStorage, error) {
	storages, err := StoreBundles([]*pathbuilder.Bundle{bundle}, index, engine)
	if err != nil {
		return nil, err
	}
	return storages[0], err
}

// StoreBundles is like StoreBundle, but takes multiple bundles
func StoreBundles(bundles []*pathbuilder.Bundle, index *Index, engine BundleEngine) ([]BundleStorage, error) {
	context := &Context{
		Index:  index,
		Engine: engine,
	}
	context.Open()

	storages := make([]BundleStorage, len(bundles))
	for i := range storages {
		storages[i] = context.Store(bundles[i])
	}
	err := context.Wait()

	return storages, err
}

// Context represents a context to extract bundle data from index into storages.
//
// A Context must be opened, and eventually waited on.
// See [Open] and [Close].
type Context struct {
	Index  *Index
	Engine BundleEngine

	errOnce sync.Once
	err     error

	extractWait  sync.WaitGroup // waiting on extracting entities in all bundles
	childAddWait sync.WaitGroup // loading child entities wait
}

// Open opens this context, and signals that multiple calls to Store() may follow.
//
// Multiple calls to Open are invalid.
func (context *Context) Open() {
	context.extractWait.Add(1)
}

// Wait signals this context that no more bundles will be loaded.
// And then waits for all bundle extracting to finish.
//
// Multiple calls to Wait() are invalid.
func (context *Context) Wait() error {
	context.extractWait.Done()
	context.extractWait.Wait()
	context.childAddWait.Wait()
	return context.err
}

// reportError stores an error in this context
// if error is non-nil, returns true.
func (context *Context) reportError(err error) bool {
	if err == nil {
		return false
	}
	context.errOnce.Do(func() {
		context.err = err
	})
	return true
}

// Store creates a new Storage for the given bundle and schedules entities to be loaded.
// May only be called between calls [Open] and [Wait].
//
// Any error that occurs is returned only by Wait.
func (context *Context) Store(bundle *pathbuilder.Bundle) BundleStorage {
	context.extractWait.Add(1)

	// create a new context
	storage, err := context.Engine(bundle)
	if context.reportError(err) {
		context.extractWait.Done()
		return nil
	}

	go func() {
		defer context.extractWait.Done()

		// determine the index of the URI within the paths describing this bundle
		// this is the length of the parent path, or zero (if it does not exist).
		var entityURIIndex int
		if bundle.Parent != nil {
			entityURIIndex = len(bundle.Group.PathArray) / 2
		}

		// stage 1: load the entities themselves
		var err error
		for path := range extractPath(bundle.Group, context.Index, &err) {
			nodes, err := path.Nodes()
			if context.reportError(err) {
				return
			}
			storage.Add(nodes[entityURIIndex], nodes)
		}
		if context.reportError(err) {
			return
		}

		// stage 2: fill all the fields
		for _, field := range bundle.Fields() {
			context.extractWait.Add(1)
			go func(field pathbuilder.Field) {
				defer context.extractWait.Done()

				var err error
				for path := range extractPath(field.Path, context.Index, &err) {
					nodes, err := path.Nodes()
					context.reportError(err)

					datum, hasDatum, err := path.Datum()
					context.reportError(err)

					if !hasDatum && len(nodes) > 0 {
						datum = nodes[len(nodes)-1]
					}
					uri := nodes[entityURIIndex]

					err = storage.AddFieldValue(uri, field.ID, datum, nodes)
					if err != storages.ErrNoEntity {
						context.reportError(err)
					}
				}
				context.reportError(err)
			}(field)
		}

		// stage 3: read child paths
		cstorages := make([]BundleStorage, len(bundle.ChildBundles))
		for i, bundle := range bundle.ChildBundles {
			cstorages[i] = context.Store(bundle)
			if cstorages[i] == nil {
				// creating the storage has failed, so we don't need to continue
				// and we can return immediatly.
				return
			}

			err := storage.RegisterChildStorage(bundle.Group.ID, cstorages[i])
			context.reportError(err)
		}

		context.childAddWait.Add(len(cstorages))

		// stage 4: register all the child entities
		go func() {
			context.extractWait.Wait()

			for i, cstorage := range cstorages {
				go func(cstorage BundleStorage, bundle *pathbuilder.Bundle) {
					defer context.childAddWait.Done()
					defer cstorage.Close()

					var err error
					for child := range cstorage.Get(entityURIIndex, &err) {
						err := storage.AddChild(child.Parent, bundle.Group.ID, child.URI)
						if err != storages.ErrNoEntity {
							context.reportError(err)
						}
					}
					context.reportError(err)
				}(cstorage, bundle.ChildBundles[i])
			}
		}()
	}()

	return storage
}

const (
	debugLogAllPaths = false   // turn this on to log all paths being queried
	datatypeEmpty    = "empty" // a datatype being recalled as "empty"
)

var debugLogID int64 // id of the current log id

// extractPath extracts values for a single path from the index.
// The returned channel is never nil.
//
// Any values found along the path are written to the returned channel which is then closed.
// If an error occurs, it is written to errDst before the channel is closed.
func extractPath(path pathbuilder.Path, index *Index, errDst *error) (c <-chan Path) {
	// if we return a nil channel, we actually want to returned a closed channel.
	// so that the caller can always safely iterate over it.
	defer func() {
		if c == nil {
			out := make(chan Path)
			close(out)
			c = out
		}
	}()

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

	set, err := index.PathsStarting(wisski.Type, URI(uris[0]))
	if err != nil {
		*errDst = err
		return nil
	}
	if debugLogAllPaths {
		size, err := set.Size()
		if err != nil {
			*errDst = err
			return nil
		}
		log.Println(debugID, uris[0], size)
	}

	for i := 1; i < len(uris); i++ {
		if i%2 == 0 {
			if err := set.Ending(wisski.Type, URI(uris[i])); err != nil {
				*errDst = err
				return nil
			}
		} else {
			if err := set.Connected(URI(uris[i])); err != nil {
				*errDst = err
				return nil
			}
		}

		if debugLogAllPaths {
			size, err := set.Size()
			if err != nil {
				*errDst = err
				return nil
			}
			log.Println(debugID, uris[i], size)
		}
	}

	return set.Paths(errDst)
}
