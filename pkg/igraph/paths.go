package igraph

import (
	"fmt"
	"strings"
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/lstream"
)

// Paths represents a set of paths in a related GraphIndex.
// It implements a very simple sparql-like query engine.
//
// A Paths object is stateful.
// A Paths object should only be created from a GraphIndex; the zero value is invalid.
// It can be further refined using the [Connected] and [Ending] methods.
type Paths[Label comparable, Datum any] struct {
	index      *IGraph[Label, Datum]
	predicates []imap.ID

	elements lstream.Stream[element]
	size     int
}

/*
// PathsWithPredicate creates a [PathSet] that represents all two-element paths
// connected by an edge with the given predicate.
func (index *IGraph[Label, Datum]) PathsWithPredicate(predicate Label) *Paths[Label, Datum] {
	p := index.labels.Forward(predicate)

	soIndex := make([]imap.ID, 0)
	index.psoIndex.Fetch(p, func(b imap.ID) {
		soIndex = append(soIndex, b)
	})

	query := index.newQuery(soIndex)
	query.expand(p)
	return query
}
*/

// PathsStarting creates a new [PathSet] that represents all one-element paths
// starting at a vertex which is connected to object with the given predicate
func (index *IGraph[Label, Datum]) PathsStarting(predicate, object Label) *Paths[Label, Datum] {
	p := index.labels.Forward(predicate)
	o := index.labels.Forward(object)

	query := index.newQuery(func(sender chan<- element) {
		index.posIndex.Fetch2(p, o, func(s imap.ID) {
			sender <- element{
				Node:   s,
				Parent: nil,
			}
		})
	})
	return query
}

// newQuery creates a new Query object that contains nodes with the given ids
func (index *IGraph[URI, Datum]) newQuery(source func(sender chan<- element)) (q *Paths[URI, Datum]) {
	q = &Paths[URI, Datum]{
		index:    index,
		elements: lstream.New(source),
		size:     -1,
	}
	return q
}

// Connected extends the sets of in this PathSet by those which
// continue the existing paths using an edge labeled with predicate.
func (set *Paths[Label, Datum]) Connected(predicate Label) {
	p := set.index.labels.Forward(predicate)
	set.predicates = append(set.predicates, p)
	set.expand(p)
}

// expand expands the nodes in this query by adding a link to each element found in the index
func (set *Paths[URI, Datum]) expand(p imap.ID) {
	set.elements = lstream.Pipe(set.elements, func(subject element, sender chan<- element) {
		set.index.psoIndex.Fetch2(p, subject.Node, func(object imap.ID) {
			sender <- element{
				Node:   object,
				Parent: &subject,
			}
		})
	})
	set.size = -1
}

// Ending restricts this set of paths to those that end in a node
// which is connected to object via predicate.
func (set *Paths[URI, Datum]) Ending(predicate URI, object URI) {
	p := set.index.labels.Forward(predicate)
	o := set.index.labels.Forward(object)
	set.restrict(p, o)
}

// restrict restricts the set of nodes by those mapped in the index
func (set *Paths[URI, Datum]) restrict(p, o imap.ID) {
	set.elements = lstream.Pipe(set.elements, func(subject element, sender chan<- element) {
		if set.index.posIndex.Has(p, o, subject.Node) {
			sender <- subject
		}
	})
	set.size = -1
}

// Size returns the number of elements in this path.
//
// NOTE(twiesing): This potentially takes a lot of memory, because we need to expand the stream.
func (set *Paths[Label, Datum]) Size() int {
	if set.size != -1 {
		return set.size
	}
	all := lstream.Drain(set.elements)
	set.size = len(all)
	set.elements = lstream.NewConcrete(all)
	return set.size
}

// Paths returns the set of paths in this PathSet.
// Paths may only be called once, and invalidates the PathSet.
func (set *Paths[Label, Datum]) Paths() (paths []Path[Label, Datum]) {
	for element := range lstream.Channel(set.elements) {
		paths = append(paths, Path[Label, Datum]{
			index:   set.index,
			nodeIDs: element.nodes(),
			edgeIDs: set.predicates,
		})
	}
	return paths
}

// element represents an element of a path
type element struct {
	// node this path ends at
	Node imap.ID

	// previous element of this path (if any)
	Parent *element
}

// nodes returns the nodes contained in this path
func (elem *element) nodes() []imap.ID {
	// create a new slice for the nodes
	slice := []imap.ID{}
	for {
		slice = append(slice, elem.Node)
		elem = elem.Parent
		if elem == nil {
			break
		}
	}

	// reverse the elements!
	for i := len(slice)/2 - 1; i >= 0; i-- {
		opp := len(slice) - 1 - i
		slice[i], slice[opp] = slice[opp], slice[i]
	}

	// and return them
	return slice
}

// Path represents a path inside a GraphIndex
type Path[Label comparable, Datum any] struct {
	// index is the index this Path belonges to
	index *IGraph[Label, Datum]

	nodeIDs   []imap.ID
	nodesOnce sync.Once
	nodes     []Label
	hasDatum  bool
	datum     Datum

	edgeIDs   []imap.ID
	edgesOnce sync.Once
	edges     []Label
}

// Nodes returns the nodes this path consists of, in order.
func (path *Path[Label, Datum]) Nodes() []Label {
	path.processNodes()
	return path.nodes
}

// Node returns the label of the node at the given index of path.
func (path *Path[Label, Datum]) Node(index int) Label {
	switch {
	case len(path.nodes) > index:
		// already computed!
		return path.nodes[index]
	case index >= len(path.nodeIDs):
		// path does not exist
		var label Label
		return label
	case index == len(path.nodeIDs)-1:
		// check if the last element has data associated with it
		last := path.nodeIDs[len(path.nodeIDs)-1]
		if path.index.data.Has(last) {
			var label Label
			return label
		}
		fallthrough
	default:
		// return the index
		return path.index.labels.Reverse(path.nodeIDs[index])
	}
}

// Datum returns the datum attached to the last node of this path, if any.
func (path *Path[Label, Datum]) Datum() (datum Datum, ok bool) {
	path.processNodes()
	return path.datum, path.hasDatum
}

func (path *Path[Label, Datum]) processNodes() {
	path.nodesOnce.Do(func() {
		if len(path.nodeIDs) == 0 {
			return
		}
		// split off the last value as a datum (if any)
		last := path.nodeIDs[len(path.nodeIDs)-1]
		path.datum, path.hasDatum = path.index.data.Get(last)
		if path.hasDatum {
			path.nodeIDs = path.nodeIDs[:len(path.nodeIDs)-1]
		}

		// turn the nodes into a set of labels
		path.nodes = make([]Label, len(path.nodeIDs))
		for j, label := range path.nodeIDs {
			path.nodes[j] = path.index.labels.Reverse(label)
		}
	})
}

// Edges returns the labels of the edges this path consists of.
func (path *Path[Label, Datum]) Edges() []Label {
	path.processEdges()
	return path.edges
}

func (path *Path[Label, Datum]) processEdges() {
	path.edgesOnce.Do(func() {
		path.edges = make([]Label, len(path.edgeIDs))
		for j, label := range path.edgeIDs {
			path.edges[j] = path.index.labels.Reverse(label)
		}
	})
}

// String turns this result into a string
func (result *Path[URI, Datum]) String() string {
	var builder strings.Builder

	result.processNodes()
	result.processEdges()

	for i, edge := range result.edges {
		fmt.Fprintf(&builder, "%v %v ", result.nodes[i], edge)
	}

	if len(result.nodes) > 0 {
		fmt.Fprintf(&builder, "%v", result.nodes[len(result.nodes)-1])
	}
	if result.hasDatum {
		fmt.Fprintf(&builder, " %#v", result.datum)
	}
	return builder.String()
}
