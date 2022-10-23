// Package exporter provides facilities for converting data from nquads into relational-like structures
package exporter

import (
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Entity represents a WissKI Entity that holds information about a specific entity
type Entity struct {
	URI       string   // URI of this entity
	Path      []string // the path of this entity
	parentURI string   // if applicable, the id of the parent entity

	Fields   map[string][]FieldValue // values for specific fields
	Children map[string][]Entity     // child paths for specific entities
}

// FieldValue represents the value of a specific field
type FieldValue struct {
	Path  []string
	Value any
}

// Triples returns the Triples belonging to this field Value
func (value FieldValue) Triples(field pathbuilder.Field) [][3]string {
	triples := make([][3]string, 0)
	for i, path := range field.PathArray {
		if i%2 == 0 { // rdf type
			triples = append(triples, [3]string{
				value.Path[i/2],
				rdfType,
				path,
			})
		} else { // connected to next element
			triples = append(triples, [3]string{
				value.Path[(i-1)/2],
				path,
				value.Path[((i-1)/2)+1],
			})
		}
	}
	return triples
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

	// first create all the entities
	// and store an index of which index they have
	entities := make([]Entity, len(uris))
	lookup := make(map[string]int, len(uris))
	for i, id := range uris {
		entities[i].Path = id.Nodes()
		uri := entities[i].Path[entityURIIndex]
		lookup[uri] = i

		// store the entity URI
		// and optionally the parentURI
		entities[i].URI = uri
		if parentIndex >= 0 {
			entities[i].parentURI = entities[i].Path[parentIndex]
		}

		// prepare maps for children and fields
		entities[i].Fields = make(map[string][]FieldValue, len(fields))
		entities[i].Children = make(map[string][]Entity, len(cPaths))
	}

	// iterate over all of the fields and store the field values
	for i, fieldPath := range fieldPaths {
		field := fields[i]
		fieldID := field.ID

		// determine the cardinality of the field
		cardinality := field.Cardinality
		if cardinality <= 0 {
			cardinality = 0
		}

		// and pre-allocate an array of the given size for it
		for i := range entities {
			entities[i].Fields[fieldID] = make([]FieldValue, 0, cardinality)
		}

		// store the actual field values
		for _, fPath := range fieldPath {
			nodes := fPath.Nodes()
			datum, _ := fPath.Datum()
			uri := nodes[entityURIIndex]

			// append the new field value!
			entities[lookup[uri]].Fields[fieldID] = append(
				entities[lookup[uri]].Fields[fieldID],
				FieldValue{
					Path:  nodes,
					Value: datum,
				},
			)
		}
	}

	// iterate through all of the child paths
	for i, child := range cPaths {
		bundleID := cBundles[i].Group.ID
		for _, entity := range child {
			index, ok := lookup[entity.parentURI]
			if !ok {
				// if there isn't a parent with this ID stuff went wrong
				// so we don't deal with it for now
				continue
			}
			entities[index].Children[bundleID] = append(entities[index].Children[bundleID], entity)
		}
	}

	// and return them!
	return entities
}
