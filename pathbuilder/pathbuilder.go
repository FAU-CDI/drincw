// Package pathbuilder contains the pathbuilder
package pathbuilder

import (
	"sort"
)

type Pathbuilder map[string]*Bundle

// Bundles returns a list of top-level bundles contained in this BundleDict
func (pb Pathbuilder) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(pb))
	for _, bundle := range pb {
		if !bundle.Toplevel() {
			continue
		}
		bundles = append(bundles, bundle)
	}
	sort.SliceStable(bundles, func(i, j int) bool {
		iBundle := bundles[i]
		jBundle := bundles[j]

		if iBundle.Group.Weight < jBundle.Group.Weight {
			return true
		}

		if iBundle.Group.Weight == jBundle.Group.Weight {
			return iBundle.importOrder < jBundle.importOrder
		}

		return false
	})
	return bundles
}

func (pb Pathbuilder) Get(BundleID string) *Bundle {
	if BundleID == "" {
		return nil
	}
	bundle, ok := pb[BundleID]
	if !ok {
		return nil
	}
	return bundle
}

// GetOrCreate creates a new bundle with the given id
func (pb Pathbuilder) GetOrCreate(BundleID string) *Bundle {
	if BundleID == "" {
		return nil
	}
	bundle, ok := pb[BundleID]
	if !ok {
		bundle = new(Bundle)
		bundle.importOrder = len(pb)
		pb[BundleID] = bundle
	}
	return bundle
}
