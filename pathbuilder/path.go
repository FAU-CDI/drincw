// Package Pathbuilder defines Pathbuilder
package pathbuilder

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

// MakeCardinality returns the cardinality to use for a call to make()
func (p Path) MakeCardinality() int {
	if p.Cardinality < 0 {
		return 0
	}
	return p.Cardinality
}

// MachineName returns the machine name of this path
func (p Path) MachineName() string {
	if p.IsGroup {
		return p.Bundle
	} else {
		return p.Field
	}
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
