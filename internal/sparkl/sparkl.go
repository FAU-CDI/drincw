// Package sparkl implements a very primitive graph index
package sparkl

import (
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type (
	Entity = wisski.Entity
	URI    = wisski.URI

	BundleStorage = storages.BundleStorage
	BundleEngine  = storages.BundleEngine

	Engine       = igraph.Engine[URI, any]
	MemoryEngine = igraph.MemoryEngine[URI, any]

	Index = igraph.IGraph[URI, any] // Index represents an index of a RDF Graph
	Paths = igraph.Paths[URI, any]  // Set of Paths inside the index
	Path  = igraph.Path[URI, any]   // Singel Path in the index
)

// Predicates represent special predicates
type Predicates struct {
	SameAs    []URI
	InverseOf []URI
}

// ParsePredicateString parses a value of comma-seperate value into a list of URIs
func ParsePredicateString(target *[]URI, value string) {
	if value == "" {
		*target = nil
		return
	}

	values := strings.Split(value, ",")
	*target = make([]URI, len(values))
	for i, value := range values {
		(*target)[i] = URI(value)
	}
}
