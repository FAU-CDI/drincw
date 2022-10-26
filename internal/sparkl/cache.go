package sparkl

import (
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Cache represents an easily accessible cache of WissKIObjects.
// It is held entirely in memory.
type Cache struct {
	// mappings from bundle id to contained entities
	BEIndex map[string][]Entity

	// names of all bundles
	BundleNames []string

	// SameAs map between entities
	SameAs map[URI]URI

	// Inverse of the SameAsMap
	Alias map[URI][]URI

	// Lookup from canonical entitiy URIs to indexes in the corresponding BIIndex
	BIIndex map[string]map[URI]int

	// index from entities into bundles
	EBIndex map[URI]string
}

// TODO: Do we want to use an IMap here?

// NewCache creates a new cache from a bundle-entity-map
func NewCache(Data map[string][]Entity, SameAs map[URI]URI) (c Cache) {
	// store the bundle-entity index
	c.BEIndex = Data
	c.BIIndex = make(map[string]map[URI]int, len(c.BEIndex))
	c.EBIndex = make(map[URI]string)
	for bundle, entities := range c.BEIndex {
		c.BIIndex[bundle] = make(map[URI]int, len(entities))
		for i, entity := range entities {
			c.BIIndex[bundle][entity.URI] = i
			c.EBIndex[entity.URI] = bundle
		}
	}

	c.BundleNames = maps.Keys(c.BEIndex)
	slices.Sort(c.BundleNames)

	// setup same-as and same-as-in
	c.SameAs = SameAs
	c.Alias = make(map[URI][]URI, len(c.SameAs))
	for alias, canon := range c.SameAs {
		c.Alias[canon] = append(c.Alias[canon], alias)
	}

	return c
}

// Canonical returns the canonical version of the given uri
func (c Cache) Canonical(uri URI) URI {
	if canon, ok := c.SameAs[uri]; ok {
		return canon
	}
	return uri
}

// Aliases returns the Aliases of the given URI, excluding itself
func (c Cache) Aliases(uri URI) []URI {
	return c.Alias[uri]
}

// Bundle returns the bundle of the given uri, if any
func (c Cache) Bundle(uri URI) (string, bool) {
	bundle, ok := c.EBIndex[c.Canonical(uri)]
	return bundle, ok
}

// Entity looks up the given entity
func (c Cache) Entity(uri URI, bundle string) (*Entity, bool) {
	index, ok := c.BIIndex[bundle][c.Canonical(uri)]
	if !ok {
		return nil, false
	}
	return &c.BEIndex[bundle][index], true
}
