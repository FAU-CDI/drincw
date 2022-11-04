package igraph

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/iterator"
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

	elements iterator.Iterator[element]
	size     int
}

// PathsStarting creates a new [PathSet] that represents all one-element paths
// starting at a vertex which is connected to object with the given predicate
func (index *IGraph[Label, Datum]) PathsStarting(predicate, object Label) (*Paths[Label, Datum], error) {
	p, err := index.labels.Forward(predicate)
	if err != nil {
		return nil, err
	}

	o, err := index.labels.Forward(object)
	if err != nil {
		return nil, err
	}

	return index.newQuery(func(sender iterator.Generator[element]) {
		err := index.posIndex.Fetch(p, o, func(s imap.ID, l imap.ID) error {
			if sender.Yield(element{
				Node:       s,
				IndexLabel: l,
				Parent:     nil,
			}) {
				return errAborted
			}
			return nil
		})

		if err != errAborted {
			sender.YieldError(err)
		}
	}), nil
}

// newQuery creates a new Query object that contains nodes with the given ids
func (index *IGraph[URI, Datum]) newQuery(source func(sender iterator.Generator[element])) (q *Paths[URI, Datum]) {
	q = &Paths[URI, Datum]{
		index:    index,
		elements: iterator.New(source),
		size:     -1,
	}
	return q
}

// Connected extends the sets of in this PathSet by those which
// continue the existing paths using an edge labeled with predicate.
func (set *Paths[Label, Datum]) Connected(predicate Label) error {
	p, err := set.index.labels.Forward(predicate)
	if err != nil {
		return err
	}
	set.predicates = append(set.predicates, p)
	return set.expand(p)
}

var errAborted = errors.New("paths: aborted")

// expand expands the nodes in this query by adding a link to each element found in the index
func (set *Paths[URI, Datum]) expand(p imap.ID) error {
	set.elements = iterator.Pipe(set.elements, func(subject element, sender iterator.Generator[element]) (stop bool) {
		err := set.index.psoIndex.Fetch(p, subject.Node, func(object imap.ID, l imap.ID) error {
			if sender.Yield(element{
				Node:       object,
				IndexLabel: l,
				Parent:     &subject,
			}) {
				return errAborted
			}
			return nil
		})

		if err != errAborted {
			sender.YieldError(err)
		}
		return err != nil && err != errAborted
	})
	set.size = -1
	return nil
}

// Ending restricts this set of paths to those that end in a node
// which is connected to object via predicate.
func (set *Paths[URI, Datum]) Ending(predicate URI, object URI) error {
	p, err := set.index.labels.Forward(predicate)
	if err != nil {
		return err
	}
	o, err := set.index.labels.Forward(object)
	if err != nil {
		return err
	}
	return set.restrict(p, o)
}

// restrict restricts the set of nodes by those mapped in the index
func (set *Paths[URI, Datum]) restrict(p, o imap.ID) error {
	set.elements = iterator.Pipe(set.elements, func(subject element, sender iterator.Generator[element]) bool {
		has, err := set.index.posIndex.Has(p, o, subject.Node)
		if err != nil {
			sender.YieldError(err)
			return true
		}
		if !has {
			return false
		}
		return sender.Yield(subject)
	})
	set.size = -1
	return nil
}

// Size returns the number of elements in this path.
//
// NOTE(twiesing): This potentially takes a lot of memory, because we need to expand the stream.
func (set *Paths[Label, Datum]) Size() (int, error) {
	if set.size != -1 {
		return set.size, nil
	}

	// we don't know the size, so we need to fully expand it
	all, err := iterator.Drain(set.elements)
	if err != nil {
		return 0, err
	}
	set.size = len(all)
	set.elements = iterator.NewFromElements(all)
	return set.size, nil
}

// Paths returns an iterator over paths contained in this Paths.
// It may only be called once, afterwards further calls may be invalid.
func (set *Paths[Label, Datum]) Paths() iterator.Iterator[Path[Label, Datum]] {
	return iterator.Map(set.elements, func(element element) Path[Label, Datum] {
		return Path[Label, Datum]{
			index:   set.index,
			nodeIDs: element.nodes(),
			edgeIDs: set.predicates,
		}
	})
}

// element represents an element of a path
type element struct {
	// node this path ends at
	Node imap.ID

	// label this node had in the index (if applicable)
	IndexLabel imap.ID

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

	errNodes  error
	nodeIDs   []imap.ID
	nodesOnce sync.Once
	nodes     []Label
	hasDatum  bool
	datum     Datum

	errEdges  error
	edgeIDs   []imap.ID
	edgesOnce sync.Once
	edges     []Label
}

// Nodes returns the nodes this path consists of, in order.
func (path *Path[Label, Datum]) Nodes() ([]Label, error) {
	path.processNodes()
	return path.nodes, path.errNodes
}

// Node returns the label of the node at the given index of path.
func (path *Path[Label, Datum]) Node(index int) (Label, error) {
	switch {
	case len(path.nodes) > index:
		// already computed!
		return path.nodes[index], nil
	case index >= len(path.nodeIDs):
		// path does not exist
		var label Label
		return label, nil
	case index == len(path.nodeIDs)-1:
		// check if the last element has data associated with it
		last := path.nodeIDs[len(path.nodeIDs)-1]
		has, err := path.index.data.Has(last)
		if has || err != nil {
			var label Label
			return label, err
		}
		fallthrough
	default:
		// return the index
		return path.index.labels.Reverse(path.nodeIDs[index])
	}
}

// Datum returns the datum attached to the last node of this path, if any.
func (path *Path[Label, Datum]) Datum() (datum Datum, ok bool, err error) {
	path.processNodes()
	return path.datum, path.hasDatum, path.errNodes
}

func (path *Path[Label, Datum]) processNodes() {
	path.nodesOnce.Do(func() {
		if len(path.nodeIDs) == 0 {
			return
		}

		// split off the last value as a datum (if any)
		last := path.nodeIDs[len(path.nodeIDs)-1]
		path.datum, path.hasDatum, path.errNodes = path.index.data.Get(last)
		if path.errNodes != nil {
			return
		}
		if path.hasDatum {
			path.nodeIDs = path.nodeIDs[:len(path.nodeIDs)-1]
		}

		// turn the nodes into a set of labels
		path.nodes = make([]Label, len(path.nodeIDs))
		for j, label := range path.nodeIDs {
			path.nodes[j], path.errNodes = path.index.labels.Reverse(label)
			if path.errNodes != nil {
				return
			}
		}
	})
}

// Edges returns the labels of the edges this path consists of.
func (path *Path[Label, Datum]) Edges() ([]Label, error) {
	path.processEdges()
	return path.edges, path.errEdges
}

func (path *Path[Label, Datum]) processEdges() {
	path.edgesOnce.Do(func() {
		path.edges = make([]Label, len(path.edgeIDs))
		for j, label := range path.edgeIDs {
			path.edges[j], path.errEdges = path.index.labels.Reverse(label)
			if path.errEdges != nil {
				return
			}
		}
	})
}

// String turns this result into a string
//
// NOTE(twiesing): This is for debugging only, and ignores all errors.
// It should not be used in producion code.
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
