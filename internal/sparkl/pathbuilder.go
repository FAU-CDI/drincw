package sparkl

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// LoadPathbuilder loads all paths from the given pathbuilder
func LoadPathbuilder(pb *pathbuilder.Pathbuilder, index *Index) map[string][]Entity {
	bundles := pb.Bundles()
	storages := StoreBundles(bundles, index, storages.NewBundleSlice)

	entities := make([][]Entity, len(bundles))

	var wg sync.WaitGroup
	for i := range storages {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			storage := storages[i]
			defer storage.Close()

			for element := range storage.Get(-1) {
				entities[i] = append(entities[i], storage.Load(element.URI))
			}
		}(i)
	}
	wg.Wait()

	result := make(map[string][]Entity, len(entities))
	for i, instances := range entities {
		result[bundles[i].Group.ID] = instances
	}
	return result
}
