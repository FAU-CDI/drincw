// Package Pathbuilder defines Pathbuilder
package pathbuilder

// cspell:words pathbuilder twiesing sparql

// Path represents a single path in the Pathbuilder
type Path struct {
	ID   string // Identifier of this path
	UUID string // UUID of this path

	Weight  int  // Display Order in the frontend
	Enabled bool // Is the path enabled or not

	IsGroup bool // Is this path a group or a field?

	GroupID string // Identifier of the group this path belongs to
	Bundle  string // Identifier of the bundle this path belongs to

	Field                string // Identifier of the field this path belongs to
	FieldType            string // Actual Field Type
	FieldTypeInformative string // Field type to display to the user

	DisplayWidget   string // Widget used for display
	FormatterWidget string // Widget used for formatting

	Cardinality int // Cardinality of this path

	PathArray        []string // Paths that make up the item
	DatatypeProperty string   // Datatype property (in case of a field)
	Disamb           int      // index where the path will be disambiguated

	Name        string // Name of this path
	ShortName   string // ShortName of this path
	Description string // Description of this path
}

const DatatypeEmpty = "empty"

func (p Path) Datatype() string {
	dp := p.DatatypeProperty
	if dp == DatatypeEmpty {
		return ""
	}
	return dp
}

// Paths returns a copy of the path array and the datatype property (if any).
// It is intended to be used during building a SPARQL query pertaining to this path.
func (p Path) Paths() []string {
	uris := make([]string, 0, len(p.PathArray)+1)

	uris = append(uris, p.PathArray...)
	if datatype := p.Datatype(); !p.IsGroup && datatype != "" {
		uris = append(uris, datatype)
	}
	return uris
}

// MakeCardinality returns the cardinality to use for a call to make()
func (p Path) MakeCardinality() int {
	if p.Cardinality < 0 {
		return 0
	}
	return p.Cardinality
}

// MachineName returns the machine name of this path.
//
// The machine name is an identifier that is guaranteed to be unique among the respective bundle, but not globally.
// If a valid path, the machine name is not the empty string.
func (p Path) MachineName() (machine string) {
	// NOTE(twiesing): MachineName is used in a lot of URL contexts and database table names.
	// As such it should "look nice", meaning it should not just be a random hash.

	if p.ID != "" {
		// ID contains a nice readable name
		// so we use it first
		return p.ID
	}

	if p.UUID != "" {
		// globally unique
		return p.UUID
	}

	if p.IsGroup {
		return p.Bundle
	} else {
		return p.Field
	}
}

// Paths recursively returns all paths in this bundle
func (pb Pathbuilder) Paths() []Path {
	paths := make([]Path, 0, len(pb.bundles))
	for _, b := range pb.Bundles() {
		paths = append(paths, b.Paths()...)
	}
	return paths
}

// Paths recursively returns all paths in this in this bundle
func (bundle Bundle) Paths() []Path {
	paths := make([]Path, 0, len(bundle.ChildBundles)+len(bundle.ChildBundles)+1)

	paths = append(paths, bundle.Path)
	for _, c := range bundle.Bundles() {
		paths = append(paths, c.Paths()...)
	}
	for _, f := range bundle.Fields() {
		paths = append(paths, f.Path)
	}
	return paths
}
