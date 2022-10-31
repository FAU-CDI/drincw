package sparkl

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// LoadPathbuilder loads all paths in the given pathbuilder
func LoadPathbuilder(pb *pathbuilder.Pathbuilder, index *Index, engine storages.BundleEngine) (map[string][]Entity, error) {
	bundles := pb.Bundles()

	storages, closer, err := StoreBundles(bundles, index, engine)
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return nil, err
	}

	entities := make([][]Entity, len(bundles))

	var errOnce sync.Once
	var gErr error

	var wg sync.WaitGroup
	for i := range storages {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			storage := storages[i]
			defer storage.Close()

			uris := storage.Get(-1)
			defer uris.Close()

			for uris.Next() {
				element := uris.Datum()
				entity, err := storage.Load(element.URI)
				if err != nil {
					errOnce.Do(func() { gErr = err })
				}
				entities[i] = append(entities[i], entity)
			}
			if err := uris.Err(); err != nil {
				errOnce.Do(func() { gErr = err })
			}
		}(i)
	}
	wg.Wait()

	result := make(map[string][]Entity, len(entities))
	for i, instances := range entities {
		result[bundles[i].Group.ID] = instances
	}
	return result, gErr
}
