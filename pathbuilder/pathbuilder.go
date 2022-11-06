// Package pathbuilder provides the Pathbuilder and related classes
package pathbuilder

import (
	"sort"
)

// Pathbuilder represents a WissKI Pathbuilder
//
// A Pathbuilder consists of an order collection of bundles.
// A singular bundle can be accessed using it's identifier.
type Pathbuilder map[string]*Bundle

// Bundles returns an ordered list of bundles in this Pathbuilder
func (pb Pathbuilder) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(pb))
	for _, bundle := range pb {
		if !bundle.IsToplevel() {
			continue
		}
		bundles = append(bundles, bundle)
	}
	sort.SliceStable(bundles, func(i, j int) bool {
		iBundle := bundles[i]
		jBundle := bundles[j]

		if iBundle.Path.Weight < jBundle.Path.Weight {
			return true
		}

		if iBundle.Path.Weight == jBundle.Path.Weight {
			return iBundle.order < jBundle.order
		}

		return false
	})
	return bundles
}

// Get returns the bundle with the given id
func (pb Pathbuilder) Get(id string) *Bundle {
	if id == "" {
		return nil
	}
	bundle, ok := pb[id]
	if !ok {
		return nil
	}
	return bundle
}

// Bundle returns the bundle with the given machine name.
// If such a bundle does not exist, returns nil.
func (pb Pathbuilder) Bundle(machine string) *Bundle {
	for _, bundle := range pb {
		if bundle.MachineName() == machine {
			return bundle
		}
	}
	return nil
}

// GetOrCreate either gets or creates a bundle
func (pb Pathbuilder) GetOrCreate(id string) *Bundle {
	if id == "" {
		return nil
	}
	bundle, ok := pb[id]
	if !ok {
		bundle = new(Bundle)
		bundle.order = len(pb)
		pb[id] = bundle
	}
	return bundle
}
