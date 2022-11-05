package wisski

import (
	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

// Entity represents an Entity inside a WissKI Bundle
type Entity struct {
	URI     URI      // URI of this entity
	Path    []URI    // the path of this entity
	Triples []Triple // the triples that define this entity itself

	Fields   map[string][]FieldValue // values for specific fields
	Children map[string][]Entity     // child paths for specific entities
}

// FieldValue represents the value of a field inside an entity
type FieldValue struct {
	Path    []URI
	Triples []Triple
	Value   any
}

// Triple represents a triple of WissKI Data
type Triple = igraph.Triple[URI, any]
