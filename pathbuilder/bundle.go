package pathbuilder

// cspell:words pathbuilder toplevel

import (
	"sort"
)

// Bundle represents a class of objects
type Bundle struct {
	// Path represents the path of this object
	Path

	Parent       *Bundle   // Parent Bundle (if any)
	ChildBundles []*Bundle // Children of this Bundle
	ChildFields  []Field   // Fields in this Bundle

	order int // tracks order of this bundle within a pathbuilder
}

// EqualBundles checks if two bundles are equal.
// Bundles are equal if they are both nil, or if their machine names are equal.
func EqualBundles(left, right *Bundle) bool {
	if left == nil || right == nil {
		return left == right
	}

	return left.MachineName() == right.MachineName()
}

// Field returns the field with the given id.
// if the field does not exist, it returns the empty field.
func (bundle Bundle) Field(id string) Field {
	for _, f := range bundle.ChildFields {
		if f.ID == id {
			return f
		}
	}
	return Field{}
}

// Bundle returns the bundle with the given machine name.
// If such a bundle does not exists, returns nil.
func (bundle Bundle) Bundle(machine string) *Bundle {
	for _, b := range bundle.ChildBundles {
		if b.MachineName() == machine {
			return b
		}
	}
	return nil
}

// IsToplevel checks if this bundle is toplevel
func (bundle Bundle) IsToplevel() bool {
	return bundle.Parent == nil
}

// Bundles returns an ordered list of child bundles.
// Bundles are ordered by their weight.
func (bundle Bundle) Bundles() []*Bundle {
	// TODO: can we cache the ordering?
	children := make([]*Bundle, len(bundle.ChildBundles))
	copy(children, bundle.ChildBundles)
	sort.SliceStable(children, func(i, j int) bool {
		return children[i].Path.Weight < children[j].Path.Weight
	})
	return children
}

// Fields
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
