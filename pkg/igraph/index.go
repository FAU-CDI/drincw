package igraph

import (
	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"golang.org/x/exp/slices"
)

// IGraph represents a searchable index of a directed labeled graph with optionally attached Data.
//
// Labels are used for nodes and edges.
// This means that the graph is defined by triples of the form (subject Label, predicate Label, object Label).
// See [AddTriple].
//
// Datum is used for data associated with the specific nodes.
// See [AddDatum].
//
// The zero value represents an empty index, but is otherwise not ready to be used.
// To fill an index, it first needs to be [Reset], and then [Finalize]d.
//
// IGraph may not be modified concurrently, however it is possible to run several queries concurrently.
type IGraph[Label comparable, Datum any] struct {
	labels imap.IMap[Label]

	// data holds mappings between internal IDs and data
	data map[imap.ID]Datum

	// the triple indexes, forward and backward
	psoIndex map[imap.ID]map[imap.ID][]imap.ID
	posIndex map[imap.ID]map[imap.ID]map[imap.ID]struct{}
}

// TripleCount returns the total number of (distinct) triples in this graph.
// Triples which have been identified will only count once.
func (index *IGraph[Label, Datum]) TripleCount() (count int64) {
	if index == nil {
		return 0
	}
	for _, soIndex := range index.psoIndex {
		for _, oList := range soIndex {
			count += int64(len(oList))
		}
	}
	return
}

// Reset resets this index and prepares all internal structures for use.
func (index *IGraph[Label, Datum]) Reset() {
	index.labels.Reset()

	index.data = make(map[imap.ID]Datum)
	index.psoIndex = make(map[imap.ID]map[imap.ID][]imap.ID)
	index.posIndex = make(map[imap.ID]map[imap.ID]map[imap.ID]struct{})
}

// AddTriple inserts a subject-predicate-object triple into the index.
// Adding a triple more than once has no effect.
//
// Reset must have been called, or this function may panic.
// After all Add operations have finished, Finalize must be called.
func (index *IGraph[Label, Datum]) AddTriple(subject, predicate, object Label) {
	index.insert(index.labels.Add(subject), index.labels.Add(predicate), index.labels.Add(object))
}

// AddData inserts a subject-predicate-data triple into the index.
// Adding multiple items to a specific subject with a specific predicate is supported.
//
// Reset must have been called, or this function may panic.
// After all Add operations have finished, Finalize must be called.
func (index *IGraph[Label, Datum]) AddData(subject, predicate Label, object Datum) {
	o := index.labels.Next()
	index.data[o] = object
	index.insert(index.labels.Add(subject), index.labels.Add(predicate), o)
}

// insert inserts the provided (subject, predicate, object) ids into the graph
func (index *IGraph[Label, Datum]) insert(subject, predicate, object imap.ID) {
	// setup the predicate-subject-object index
	if index.psoIndex[predicate] == nil {
		index.psoIndex[predicate] = make(map[imap.ID][]imap.ID, 1)
	}
	index.psoIndex[predicate][subject] = append(index.psoIndex[predicate][subject], object)

	// setup the predicate-object-subject index
	if index.posIndex[predicate] == nil {
		index.posIndex[predicate] = make(map[imap.ID]map[imap.ID]struct{}, 1)
	}
	if index.posIndex[predicate][object] == nil {
		index.posIndex[predicate][object] = make(map[imap.ID]struct{}, 1)
	}
	index.posIndex[predicate][object][subject] = struct{}{}
}

// Identify identifies the left and right labels.
// The identify function must run before any Add calls are made.
func (index *IGraph[Label, Datum]) Identify(left, right Label) {
	index.labels.Identify(left, right)
}

// IdentifyMap returns the canonical names of labels
func (index *IGraph[Label, Datum]) IdentityMap() map[Label]Label {
	return index.labels.IdentifyMap()
}

// Finalize finalizes any adding operations into this graph.
//
// Finalize must be called before any query is performed,
// but after any calls to the Add* methods.
// Calling finalize multiple times is valid.
func (index *IGraph[Label, Datum]) Finalize() {
	for pred := range index.psoIndex {
		for sub := range index.psoIndex[pred] {
			slices.SortFunc(index.psoIndex[pred][sub], func(a, b imap.ID) bool { return a.Less(b) })
			index.psoIndex[pred][sub] = slices.Compact(index.psoIndex[pred][sub])
		}
	}
}
