package sparkl

import (
	"sync"

	"github.com/FAU-CDI/drincw/internal/sparkl/exporter"
	"github.com/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/FAU-CDI/drincw/internal/wisski"
	"github.com/FAU-CDI/drincw/pathbuilder"
)

// Export loads all top-level paths from the given path-builder from the index into the given engine.
// Afterwards it is exported into the given exporter.
func Export(pb *pathbuilder.Pathbuilder, index *Index, engine storages.BundleEngine, exporter exporter.Exporter) error {
	bundles := pb.Bundles()

	storages, closer, err := StoreBundles(bundles, index, engine)
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return err
	}

	var errOnce sync.Once
	var gErr error

	var wg sync.WaitGroup
	for i := range storages {
		wg.Add(1)
		go (func(i int) {
			defer wg.Done()

			err := func() (e error) {
				storage := storages[i]
				bundle := bundles[i]
				defer storage.Close()

				// count the number of elements
				count, err := storage.Count()
				if err != nil {
					return err
				}

				// start the exporter
				if err := exporter.Begin(bundle, count); err != nil {
					errOnce.Do(func() { gErr = err })
					return err
				}

				// make sure it is also closed
				defer func() {
					err := exporter.End(bundle)
					if e == nil {
						e = err
					}
				}()

				// load uris from storage
				uris := storage.Get(-1)
				defer uris.Close()

				// load all the entities
				for uris.Next() {
					element := uris.Datum()
					entity, err := storage.Load(element.URI)
					if err != nil {
						return err
					}
					if err := exporter.Add(bundle, &entity); err != nil {
						return err
					}
				}

				// and return it
				return uris.Err()
			}()

			if err != nil {
				errOnce.Do(func() { gErr = err })
			}
		})(i)
	}
	wg.Wait()

	// close the exporter
	{
		err := exporter.Close()
		if gErr == nil {
			gErr = err
		}
	}

	return gErr
}

// LoadPathbuilder loads all paths in the given pathbuilder
func LoadPathbuilder(pb *pathbuilder.Pathbuilder, index *Index, engine storages.BundleEngine) (map[string][]Entity, error) {
	mp := exporter.Map{
		Data: make(map[string][]wisski.Entity, len(pb.Bundles())),
	}
	err := Export(pb, index, engine, &mp)
	if err != nil {
		return nil, err
	}
	return mp.Data, nil
}
