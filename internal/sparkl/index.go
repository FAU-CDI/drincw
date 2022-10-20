package sparkl

import (
	"golang.org/x/exp/slices"
)

// GraphIndex represents a searchable index of a directed labeled graph with optionally attached Data.
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
// GraphIndex may not be modified concurrently, however it is possible to run several queries concurrently.
type GraphIndex[Label comparable, Datum any] struct {
	// forward and reverse hold mappings between the Labels and indexes.
	forward map[Label]indexID
	reverse map[indexID]Label

	// data holds mappings between internal IDs and data
	data map[indexID]Datum

	last indexID // last id inserted

	// the triple indexes, forward and backward
	psoIndex map[indexID]map[indexID][]indexID
	posIndex map[indexID]map[indexID]map[indexID]struct{}
}

// TripleCount returns the total number of (distinct) triples in the index
func (index *GraphIndex[Label, Datum]) TripleCount() (count int64) {
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
func (index *GraphIndex[Label, Datum]) Reset() {
	index.forward = make(map[Label]indexID)
	index.reverse = make(map[indexID]Label)
	index.data = make(map[indexID]Datum)
	index.psoIndex = make(map[indexID]map[indexID][]indexID)
	index.posIndex = make(map[indexID]map[indexID]map[indexID]struct{})
}

// AddTriple inserts a subject-predicate-object triple into the index.
// Adding a triple more than once has no effect.
//
// Reset must have been called, or this function may panic.
// After all Add operations have finished, Finalize must be called.
func (index *GraphIndex[Label, Datum]) AddTriple(subject, predicate, object Label) {
	index.insert(index.add(subject), index.add(predicate), index.add(object))
}

// AddData inserts a subject-predicate-data triple into the index.
// Adding multiple items to a specific subject with a specific predicate is supported.
//
// Reset must have been called, or this function may panic.
// After all Add operations have finished, Finalize must be called.
func (index *GraphIndex[Label, Datum]) AddData(subject, predicate Label, object Datum) {
	o := index.last.next()
	index.data[o] = object
	index.insert(index.add(subject), index.add(predicate), o)
}

// add returns or adds an internal mapping for the provided label
func (index *GraphIndex[Label, Datum]) add(label Label) indexID {
	if value, ok := index.forward[label]; ok {
		return value
	}

	// create a new mapping by inserting into the forward and reverse
	value := index.last.next()
	index.forward[label] = value
	index.reverse[value] = label
	return value
}

// insert inserts the provided (subject, predicate, object) ids into the graph
func (index *GraphIndex[Label, Datum]) insert(subject, predicate, object indexID) {
	// setup the predicate-subject-object index
	if index.psoIndex[predicate] == nil {
		index.psoIndex[predicate] = make(map[indexID][]indexID, 1)
	}
	index.psoIndex[predicate][subject] = append(index.psoIndex[predicate][subject], object)

	// setup the predicate-object-subject index
	if index.posIndex[predicate] == nil {
		index.posIndex[predicate] = make(map[indexID]map[indexID]struct{}, 1)
	}
	if index.posIndex[predicate][object] == nil {
		index.posIndex[predicate][object] = make(map[indexID]struct{}, 1)
	}
	index.posIndex[predicate][object][subject] = struct{}{}
}

// Finalize finalizes any adding operations into this graph.
//
// Finalize must be called before any query is performed,
// but after any calls to the Add* methods.
// Calling finalize multiple times is valid.
func (index *GraphIndex[Label, Datum]) Finalize() {
	for pred := range index.psoIndex {
		for sub := range index.psoIndex[pred] {
			index.psoIndex[pred][sub] = slices.Compact(index.psoIndex[pred][sub])
		}
	}
}

// indexID represents an item in the index
type indexID int64

// next increments this ID, and then returns a copy of the new value.
// It is the equivalent of the "++" operator.
func (i *indexID) next() indexID {
	(*i)++
	return *i
}
