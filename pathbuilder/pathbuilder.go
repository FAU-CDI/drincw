// Package pathbuilder contains the pathbuilder
package pathbuilder

import "sort"

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
		return bundles[i].Group.Weight < bundles[j].Group.Weight
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
func (pb Pathbuilder) GetOrCreate(BundleID string) *Bundle {
	if BundleID == "" {
		return nil
	}
	bundle, ok := pb[BundleID]
	if !ok {
		bundle = new(Bundle)
		pb[BundleID] = bundle
	}
	return bundle
}
