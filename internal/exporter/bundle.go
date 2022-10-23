// Package exporter provides facilities for converting data from nquads into relational-like structures
package exporter

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Entity represents a WissKI Entity that holds information about a specific entity
type Entity struct {
	URI       string // URI of this entity
	parentURI string // if applicable, the id of the parent entity

	Fields   map[string]any      // values for specific fields
	Children map[string][]Entity // child paths for specific entities
}

// Entities loads all entities for a specific bundle from the given index
func Entities(bundle *pathbuilder.Bundle, index *Index) []Entity {
	return entities(bundle, -1, index)
}

// entities implements Entities.
// parentIndex is the index of the parentURI (for nested entities)
func entities(bundle *pathbuilder.Bundle, parentIndex int, index *Index) []Entity {
	// determine the index of the URI within the paths describing this bundle
	// this is the length of the parent path, or zero (if it does not exist).
	var entityURIIndex int
	if bundle.Parent != nil {
		entityURIIndex = len(bundle.Group.PathArray) / 2
	}

	var wg sync.WaitGroup

	// scan this graph for the main path that describes this bundle
	var uris []Path
	wg.Add(1)
	go func() {
		defer wg.Done()
		uris = FromPath(bundle.Group, index)
	}()

	// scan all the paths for all of the fields
	fields := bundle.Fields()
	fieldPaths := make([][]Path, len(fields))
	wg.Add(len(fields))
	for i := range fields {
		i := i

		go func() {
			defer wg.Done()
			fieldPaths[i] = FromPath(fields[i].Path, index)
		}()
	}

	// fetch all the child entities
	cBundles := bundle.ChildBundles
	cPaths := make([][]Entity, len(bundle.Bundles()))
	wg.Add(len(cBundles))
	for i := range cBundles {
		i := i
		go func() {
			defer wg.Done()

			// store the URI of the parent in the entity URI index!
			cPaths[i] = entities(cBundles[i], entityURIIndex, index)
		}()
	}

	wg.Wait()

	// make the entities
	entities := make([]Entity, len(uris))
	lookup := make(map[string]int, len(uris))
	for i, id := range uris {
		uri := id.Node(entityURIIndex)
		lookup[uri] = i

		entities[i].URI = uri
		if parentIndex >= 0 {
			entities[i].parentURI = id.Node(parentIndex)
		}
		entities[i].Fields = make(map[string]any, len(fields))
		entities[i].Children = make(map[string][]Entity, len(cPaths))
	}

	// setup all the fields
	for i, fieldPath := range fieldPaths {
		field := fields[i].ID
		for _, fPath := range fieldPath {
			uri := fPath.Node(entityURIIndex)

			entities[lookup[uri]].Fields[field], _ = fPath.Datum()
		}
	}

	// iterate through all of the child paths
	for i, child := range cPaths {
		bundle := cBundles[i].Group.ID
		for _, entity := range child {
			index, ok := lookup[entity.parentURI]
			if !ok {
				// if there isn't a parent with this ID stuff went wrong
				// so we don't deal with it for now
				continue
			}
			entities[index].Children[bundle] = append(entities[index].Children[bundle], entity)
		}
	}

	// and return them!
	return entities
}
