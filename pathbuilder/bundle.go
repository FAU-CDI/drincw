package pathbuilder

import "sort"

type Bundle struct {
	Group Path

	Parent       *Bundle   // Parent Bundle (if any)
	ChildBundles []*Bundle // Children of this Bundle
	ChildFields  []Field   // Fields in this Bundle
}

// TopLevel checks if a bundle is toplevel
func (bundle Bundle) Toplevel() bool {
	return bundle.Parent == nil
}

// Bundles returns a list of child bundles in this Bundle
func (bundle Bundle) Bundles() []*Bundle {
	children := make([]*Bundle, len(bundle.ChildBundles))
	copy(children, bundle.ChildBundles)
	sort.SliceStable(children, func(i, j int) bool {
		return children[i].Group.Weight < children[j].Group.Weight
	})
	return children
}

func (bundle Bundle) Fields() []Field {
	fields := make([]Field, len(bundle.ChildFields))
	copy(fields, bundle.ChildFields)
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Weight < fields[j].Weight
	})
	return bundle.ChildFields
}

func (bundle Bundle) AllFields() []Field {
	fields := bundle.Fields()
	for _, bundle := range bundle.Bundles() {
		fields = append(fields, bundle.AllFields()...)
	}
	return fields
}
