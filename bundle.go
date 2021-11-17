package drincw

import (
	"sort"
)

type Bundle struct {
	Group Path

	Parent       *Bundle   // Parent Bundle (if any)
	ChildBundles []*Bundle // Children of this Bundle
	ChildFields  []Field   // Fields in this Bundle
}

func (bundle Bundle) Toplevel() bool {
	return bundle.Parent == nil
}

// Bundles returns a list of child bundles in this Bundle
func (bundle Bundle) Bundles() []*Bundle {
	children := make([]*Bundle, len(bundle.ChildBundles))
	copy(children, bundle.ChildBundles)
	SortableBundles(children).Sort()
	return children
}

func (bundle Bundle) Fields() []Field {
	fields := make([]Field, len(bundle.ChildFields))
	copy(fields, bundle.ChildFields)
	SortableFields(fields).Sort()
	return bundle.ChildFields
}

type Field struct {
	Path
}

type BundleDict map[string]*Bundle

// Bundles returns a list of top-level bundles contained in this BundleDict
func (dict BundleDict) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(dict))
	for _, bundle := range dict {
		if !bundle.Toplevel() {
			continue
		}
		bundles = append(bundles, bundle)
	}
	SortableBundles(bundles).Sort()
	return bundles
}

func (dict BundleDict) Fields() []Field {
	return nil
}

func (dict BundleDict) Get(BundleID ZeroString) *Bundle {
	id := string(BundleID)
	if id == "" {
		return nil
	}
	bundle, ok := dict[id]
	if !ok {
		return nil
	}
	return bundle
}
func (dict BundleDict) GetOrCreate(BundleID ZeroString) *Bundle {
	id := string(BundleID)
	if id == "" {
		return nil
	}
	bundle, ok := dict[id]
	if !ok {
		bundle = new(Bundle)
		dict[id] = bundle
	}
	return bundle
}

func (pb PathbuilderInterface) BundleDict() BundleDict {
	dict := BundleDict(map[string]*Bundle{})
	for _, path := range pb.Paths {
		if !path.Enabled {
			continue
		}

		// get the parent group
		parent := dict.GetOrCreate(path.GroupID)

		// if we don't have a group, we have a field!
		if !path.IsGroup {
			if parent == nil { // bundle-less fields shouldn't happen
				continue
			}

			parent.ChildFields = append(parent.ChildFields, Field{Path: path})
			continue
		}

		// create a new child group
		group := dict.GetOrCreate(ZeroString(path.ID))
		group.Group = path
		group.Parent = parent
		if parent != nil {
			parent.ChildBundles = append(parent.ChildBundles, group)
		}
	}
	return dict
}

type SortableBundles []*Bundle

func (bundles SortableBundles) Len() int {
	return len(bundles)
}

func (bundles SortableBundles) Swap(i, j int) {
	bundles[i], bundles[j] = bundles[j], bundles[i]
}
func (bundles SortableBundles) Less(i, j int) bool {
	return bundles[i].Group.Weight < bundles[j].Group.Weight
}

func (s SortableBundles) Sort() {
	sort.Sort(s)
}

type SortableFields []Field

func (fields SortableFields) Len() int {
	return len(fields)
}

func (fields SortableFields) Swap(i, j int) {
	fields[i], fields[j] = fields[j], fields[i]
}
func (fields SortableFields) Less(i, j int) bool {
	return fields[i].Weight < fields[j].Weight
}

func (fields SortableFields) Sort() {
	sort.Sort(fields)
}
