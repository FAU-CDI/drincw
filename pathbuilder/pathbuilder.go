// Package pathbuilder provides the Pathbuilder and related classes
package pathbuilder

import (
	"sort"
)

// Pathbuilder represents a WissKI Pathbuilder
//
// A Pathbuilder consists of an order collection of bundles.
// A singular bundle can be accessed using it's identifier.
type Pathbuilder struct {
	bundles map[string]*Bundle
}

func NewPathbuilder() Pathbuilder {
	return Pathbuilder{
		bundles: make(map[string]*Bundle),
	}
}

// Bundles returns an ordered list of main bundles in this Pathbuilder
func (pb Pathbuilder) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(pb.bundles))
	for _, bundle := range pb.bundles {
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
	bundle, ok := pb.bundles[id]
	if !ok {
		return nil
	}
	return bundle
}

// Bundle returns the main bundle with the given machine name.
// If such a bundle does not exist, returns nil.
func (pb Pathbuilder) Bundle(machine string) *Bundle {
	for _, bundle := range pb.Bundles() {
		if bundle.MachineName() == machine {
			return bundle
		}
	}
	return nil
}

// FindBundle returns the (main or nested) bundle with the given machine name.
// If such a bundle does not exist, returns nil.
func (pb Pathbuilder) FindBundle(machine string) *Bundle {
	return pb.bundles[machine]
}

// GetOrCreate either gets or creates a bundle
func (pb Pathbuilder) GetOrCreate(id string) *Bundle {
	if id == "" {
		return nil
	}
	bundle, ok := pb.bundles[id]
	if !ok {
		bundle = new(Bundle)
		bundle.order = len(pb.bundles)
		pb.bundles[id] = bundle
	}
	return bundle
}
