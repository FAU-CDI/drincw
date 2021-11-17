package drincw

type Path struct {
	ID      string
	Weight  int
	Enabled bool

	GroupID string
	Bundle  string

	Field     string
	FieldType string

	DisplayWidget   string
	FormatterWidget string

	Cardinality int

	FieldTypeInformative string

	PathArray []string

	DatatypeProperty string

	ShortName string
	Disamb    int

	Description string
	UUID        string

	IsGroup bool

	Name string
}

// Paths recursively returns all paths in this bundle
func (pb Pathbuilder) Paths() []Path {
	paths := make([]Path, 0, len(pb))
	for _, b := range pb.Bundles() {
		paths = append(paths, b.Paths()...)
	}
	return paths
}

// Paths recursively returns all paths in this in this bundle
func (bundle Bundle) Paths() []Path {
	paths := make([]Path, 0, len(bundle.ChildBundles)+len(bundle.ChildBundles)+1)

	paths = append(paths, bundle.Group)
	for _, c := range bundle.Bundles() {
		paths = append(paths, c.Paths()...)
	}
	for _, f := range bundle.Fields() {
		paths = append(paths, f.Path)
	}
	return paths
}

func (x XMLPath) Path() (p Path) {
	p.ID = x.ID
	p.Weight = x.Weight
	p.Enabled = bool(x.Enabled)
	p.GroupID = string(x.GroupID)
	p.Bundle = x.Bundle
	p.Field = x.Field
	p.FieldType = x.FieldType
	p.DisplayWidget = x.DisplayWidget
	p.FormatterWidget = x.FormatterWidget
	p.Cardinality = x.Cardinality
	p.FieldTypeInformative = x.FieldTypeInformative
	p.PathArray = x.PathArray
	p.DatatypeProperty = x.DatatypeProperty
	p.ShortName = x.ShortName
	p.Disamb = x.Disamb
	p.Description = x.Description
	p.UUID = x.UUID
	p.IsGroup = bool(x.IsGroup)
	p.Name = x.Name
	return
}

type Pathbuilder map[string]*Bundle

func (xml XMLPathbuilder) Pathbuilder() Pathbuilder {
	pb := Pathbuilder(map[string]*Bundle{})
	for _, path := range xml.Paths {
		if !path.Enabled {
			continue
		}

		// get the parent group
		parent := pb.GetOrCreate(string(path.GroupID))

		// if we don't have a group, we have a field!
		if !path.IsGroup {
			if parent == nil { // bundle-less fields shouldn't happen
				continue
			}

			parent.ChildFields = append(parent.ChildFields, Field{Path: path.Path()})
			continue
		}

		// create a new child group
		group := pb.GetOrCreate(path.ID)
		group.Group = path.Path()
		group.Parent = parent
		if parent != nil {
			parent.ChildBundles = append(parent.ChildBundles, group)
		}
	}
	return pb
}

// Bundles returns a list of top-level bundles contained in this BundleDict
func (pb Pathbuilder) Bundles() []*Bundle {
	bundles := make([]*Bundle, 0, len(pb))
	for _, bundle := range pb {
		if !bundle.Toplevel() {
			continue
		}
		bundles = append(bundles, bundle)
	}
	sortableBundles(bundles).Sort()
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
	sortableBundles(children).Sort()
	return children
}

func (bundle Bundle) Fields() []Field {
	fields := make([]Field, len(bundle.ChildFields))
	copy(fields, bundle.ChildFields)
	sortableFields(fields).Sort()
	return bundle.ChildFields
}

type Field struct {
	Path
}
