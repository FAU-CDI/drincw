package wisski

import "github.com/tkw1536/FAU-CDI/drincw/pathbuilder"

// Entity represents an Entity inside a WissKI Bundle
type Entity struct {
	URI  URI   // URI of this entity
	Path []URI // the path of this entity

	Fields   map[string][]FieldValue // values for specific fields
	Children map[string][]Entity     // child paths for specific entities
}

// FieldValue represents the value of a field inside an entity
type FieldValue struct {
	Path  []URI
	Value any
}

// Triples reconstructs triples that represent the field.
//
// It is not guaranteed that these triples are exactly the same triples as the original field.
// It is guaranteed that reparsing these triples results in the same field value.
func (value FieldValue) Triples(field pathbuilder.Field) [][3]URI {
	// NOTE(twiesing): We might want to re-do this
	triples := make([][3]URI, 0)
	for i, path := range field.PathArray {
		if i%2 == 0 { // rdf type
			triples = append(triples, [3]URI{
				value.Path[i/2],
				Type,
				URI(path),
			})
		} else { // connected to next element
			triples = append(triples, [3]URI{
				value.Path[(i-1)/2],
				URI(path),
				value.Path[((i-1)/2)+1],
			})
		}
	}
	return triples
}
