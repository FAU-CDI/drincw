package sparkl

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// LoadPathbuilder loads all paths from the given pathbuilder
func LoadPathbuilder(pb *pathbuilder.Pathbuilder, index *Index) map[string][]Entity {
	bundles := pb.Bundles()
	entities := make([][]Entity, len(bundles))

	var wg sync.WaitGroup
	for i := range bundles {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			for entity := range ExtractEntities(bundles[i], index, NewBundleSlice).Get() {
				entities[i] = append(entities[i], entity)
			}
		}()
	}
	wg.Wait()

	result := make(map[string][]Entity, len(entities))
	for i, instances := range entities {
		result[bundles[i].Group.ID] = instances
	}
	return result
}
