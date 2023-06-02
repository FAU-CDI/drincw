// Package dot transforms a pathbuilder into a dot file
package dot

import (
	"strings"

	"github.com/FAU-CDI/drincw/pathbuilder"
	"github.com/emicklei/dot"
)

type Options struct {
	Prefixes map[string]string // prefixes for urls to use

	IDPrefix string // force id prefixes for specific nodes

	FlatChildBundles        bool // do not create groups for each child bundle
	IndependentChildBundles bool // consider each bundle a
	CopyChildBundleNodes    bool // attempt to copy nodes which occur in unique child bundles

	BundleUseDisplayNames bool   // use display names (as opposed to machine names) for bundle labels
	ColorBundle           string // color to highlight starting points for bundles
	ColorDatatype         string // color to use for datatype property nodes
}

// NewDot creates a new graph from the given pathbuilder
func NewDot(pb pathbuilder.Pathbuilder, opts Options) *dot.Graph {
	return NewDotForBundles(opts, pb.Bundles()...)
}

// NewDot creates a new graph from the given pathbuilder
func NewDotForBundles(opts Options, bundles ...*pathbuilder.Bundle) *dot.Graph {
	g := dot.NewGraph(dot.Directed)

	prefix := opts.IDPrefix
	for _, bundle := range bundles {
		if bundle == nil || !bundle.Enabled {
			continue
		}
		opts.IDPrefix = prefix + ":::" + bundle.MachineName()
		addBundle(NewBundleSubgraph(g, bundle, opts), bundle, opts, make(map[*dot.Graph]struct{}))
	}

	return g
}

// AddBundle adds output for the given bundle to the given graph.
func addBundle(g *dot.Graph, bundle *pathbuilder.Bundle, opts Options, gs map[*dot.Graph]struct{}) {
	gs[g] = struct{}{} // add the current graph

	for _, field := range bundle.ChildFields {
		addField(g, field, bundle, opts, gs)
	}

	for _, bundle := range bundle.ChildBundles {
		addBundle(NewBundleSubgraph(g, bundle, opts), bundle, opts, gs)
	}
}

// NewBundleSubgraph adds a new subgraph for the given bundle.
// It does not fill the subgraph.
func NewBundleSubgraph(g *dot.Graph, bundle *pathbuilder.Bundle, opts Options) *dot.Graph {
	if opts.FlatChildBundles && !bundle.IsToplevel() {
		return g
	}
	b := g.Subgraph(bundle.MachineName(), dot.ClusterOption{})
	if opts.BundleUseDisplayNames {
		b.Label(bundle.Name)
	} else {
		b.Label(bundle.MachineName())
	}
	return b
}

func addField(g *dot.Graph, field pathbuilder.Field, bundle *pathbuilder.Bundle, opts Options, gs map[*dot.Graph]struct{}) {
	var prev, now dot.Node

	var maxParentID int
	if bundle != nil {
		maxParentID = len(bundle.PathArray)
		if maxParentID%2 == 1 {
			maxParentID--
		}
	}

	var reuse bool // did we re-use an existing node in the parent?

	for i := 0; i < len(field.PathArray); i += 2 {
		gs := gs // list of graphs to search for existing edges in

		prev = now // advance to the next node

		// check if we can reuse an existing node
		id := opts.NodeID(field.PathArray[i], bundle)
		now, reuse = g.Root().FindNodeById(id)

		if opts.CopyChildBundleNodes {
			reuse = reuse && i < maxParentID // only when we are not in the new bundle
		}

		if !reuse {
			label := opts.FormatID(field.PathArray[i])
			now = g.Node(id).Label(label) // create a new node
			gs = nil                      // force re-creating edges
		}

		// color the node if we didn't re-create it!
		if !reuse && i == maxParentID {
			now.Attr("fontcolor", opts.ColorBundle).Attr("color", opts.ColorBundle)
		}

		if i == 0 {
			continue
		}

		// do the edge insertion
		rel := opts.FormatID(field.PathArray[i-1])
		if anyHasEdge(g, gs, prev, now, rel) {
			continue
		}
		g.Edge(prev, now, rel)
	}

	// add a node for a datatype property
	dp := opts.FormatID(field.Datatype())
	if dp == "" || now.ID() == "" {
		return
	}

	// create a node for the field

	data := g.Node(opts.NodeID(field.MachineName(), bundle)).Label(field.Name)
	edge := g.Edge(now, data, dp)
	if opts.ColorDatatype != "" {
		data.Attr("fontcolor", opts.ColorDatatype).Attr("color", opts.ColorDatatype)
		edge.Attr("color", opts.ColorDatatype)
	}
}

func anyHasEdge(g *dot.Graph, gs map[*dot.Graph]struct{}, from, to dot.Node, label string) bool {
	if hasEdge(g, from, to, label) {
		return true
	}

	for g := range gs {
		if hasEdge(g, from, to, label) {
			return true
		}
	}
	return false
}

// hasEdge checks if from and to have an edge between them
func hasEdge(g *dot.Graph, from, to dot.Node, label string) bool {
	for _, edge := range g.FindEdges(from, to) {
		if edge.GetAttr("label") == label {
			return true
		}
	}

	return false
}

// NodeID returns the node id for a node with the given id inside the given bundle.
func (opts Options) NodeID(id string, bundle *pathbuilder.Bundle) string {
	id = opts.IDPrefix + ":::" + id
	if opts.IndependentChildBundles && bundle != nil {
		return bundle.MachineName() + ":::" + id
	}
	return opts.IDPrefix + ":::" + id
}

func (opts Options) FormatID(id string) string {

	// find the longest prefix that matches
	var name, prefix string
	for n, p := range opts.Prefixes {
		if !strings.HasPrefix(id, p) {
			continue
		}

		if len(p) > len(prefix) {
			name = n
			prefix = p
		}
	}

	// no prefix found
	if name == "" {
		return id
	}

	// apply the prefix
	return name + ":" + strings.TrimPrefix(id, prefix)
}
