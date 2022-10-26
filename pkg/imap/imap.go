package imap

// IMap holds forward and reverse mapping from Labels to IDs.
// An IMap may be read concurrently; however any operations which change internal state are not safe to access concurrently.
//
// The zero map is not ready for use; it should be initialized using a call to [Reset].
type IMap[Label comparable] struct {
	forward Storage[Label, ID]
	reverse Storage[ID, Label]

	id ID // last id inserted
}

// Reset resets this IMap to be empty.
func (mp *IMap[Label]) Reset() {
	mp.forward = make(MapStorage[Label, ID])
	mp.reverse = make(MapStorage[ID, Label])
	mp.id.Reset()
}

// Next returns a new unused id within this map
// It is always valid.
func (mp *IMap[Label]) Next() ID {
	return mp.id.Inc()
}

// Add inserts label into this IMap and returns the corresponding ID.
//
// When label (or any object marked identical to ID) already exists in this IMap, returns the corresponding ID.
func (mp *IMap[Label]) Add(label Label) (id ID) {
	id, _ = mp.AddNew(label)
	return
}

// AddNew behaves like Add, except additionally returns a boolean indiciating if the returned id existed previously.
func (mp *IMap[Label]) AddNew(label Label) (id ID, old bool) {
	// fetch the mapping (if any)
	id, old = mp.forward.Get(label)
	if old {
		return
	}

	// fetch the new id
	id = mp.id.Inc()

	// store mappings in both directions
	mp.forward.Set(label, id)
	mp.reverse.Set(id, label)

	// return the id
	return
}

// MarkIdentical marks the two labels as being identical.
// It returns the ID corresponding to the label new.
//
// Once applied, all future calls to [Forward] or [Add] with old will act as if being called by new.
// A previous ID corresponding to old (if any) is no longer valid.
//
// NOTE(twiesing): Each call to MarkIdentical potentially requires iterating over all calls that were previously added to this map.
// This is a potentially slow operation and should be avoided.
func (mp *IMap[Label]) MarkIdentical(new, old Label) (canonical ID) {
	// left and right are the same object
	canonical = mp.Add(new)
	alias, aliasIsOld := mp.AddNew(old)

	// the canonical
	if canonical == alias {
		return
	}

	// optimization: if the alias was new
	if !aliasIsOld {
		mp.forward.Set(old, canonical)
		mp.reverse.Delete(alias)
		return
	}

	// iterate over all the items
	mp.forward.Iterate(func(label Label, id ID) {
		if id != alias || label == new {
			return
		}

		mp.forward.Set(label, canonical)

		// delete the reverse mapping of the alias
		// because it cannot ever be returned
		mp.reverse.Delete(id)
	})

	return
}

// Forward returns the id corresponding to the given label.
//
// If the label is not contained in this map, the zero ID is returned.
// The zero ID is never returned for a valid id.
func (mp *IMap[Label]) Forward(label Label) ID {
	return mp.forward.GetZero(label)
}

// Reverse returns the label corresponding to the given id.
// When id is not contained in this map, the zero value of the label type is contained.
func (mp *IMap[Label]) Reverse(id ID) Label {
	return mp.reverse.GetZero(id)
}

// IdentityMap writes canonical label mappings to the given storage.
//
// Concretely a pair (L1, L2) is written to storage iff
//
//	mp.Reverse(mp.Forward(L1)) == L2 && L1 != L2
func (mp *IMap[Label]) IdentityMap(storage Storage[Label, Label]) {
	mp.forward.Iterate(func(label Label, id ID) {
		if value := mp.reverse.GetZero(id); value != label {
			storage.Set(label, value)
		}
	})
}
