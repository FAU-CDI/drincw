package imap

import "io"

// IMap holds forward and reverse mapping from Labels to IDs.
// An IMap may be read concurrently; however any operations which change internal state are not safe to access concurrently.
//
// The zero map is not ready for use; it should be initialized using a call to [Reset].
type IMap[Label comparable] struct {
	forward Storage[Label, ID]
	reverse Storage[ID, Label]

	id ID // last id inserted
}

// Reset resets this IMap to be empty, finishing any previ
func (mp *IMap[Label]) Reset(engine Engine[Label]) error {
	if err := mp.Close(); err != nil {
		return err
	}

	var err error
	var closers []io.Closer

	mp.forward, err = engine.Forward()
	if err != nil {
		return err
	}
	closers = append(closers, mp.forward)

	mp.reverse, err = engine.Reverse()
	if err != nil {
		for _, closer := range closers {
			closer.Close()
		}
		return err
	}

	mp.id.Reset()
	return nil
}

// Next returns a new unused id within this map
// It is always valid.
func (mp *IMap[Label]) Next() ID {
	return mp.id.Inc()
}

// Add inserts label into this IMap and returns the corresponding ID.
//
// When label (or any object marked identical to ID) already exists in this IMap, returns the corresponding ID.
func (mp *IMap[Label]) Add(label Label) (id ID, err error) {
	id, _, err = mp.AddNew(label)
	return
}

// AddNew behaves like Add, except additionally returns a boolean indiciating if the returned id existed previously.
func (mp *IMap[Label]) AddNew(label Label) (id ID, old bool, err error) {
	// fetch the mapping (if any)
	id, old, err = mp.forward.Get(label)
	if err != nil {
		return
	}
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
func (mp *IMap[Label]) MarkIdentical(new, old Label) (canonical ID, err error) {
	// left and right are the same object
	canonical, err = mp.Add(new)
	if err != nil {
		return canonical, err
	}
	alias, aliasIsOld, err := mp.AddNew(old)
	if err != nil {
		return canonical, err
	}

	// the canonical
	if canonical == alias {
		return
	}

	// optimization: if the alias was new
	if !aliasIsOld {
		if err := mp.forward.Set(old, canonical); err != nil {
			return canonical, err
		}
		if err := mp.reverse.Delete(alias); err != nil {
			return canonical, err
		}
		return
	}

	// iterate over all the items
	err = mp.forward.Iterate(func(label Label, id ID) error {
		if id != alias || label == new {
			return nil
		}

		if err := mp.forward.Set(label, canonical); err != nil {
			return err
		}

		// delete the reverse mapping of the alias
		// because it cannot ever be returned
		return mp.reverse.Delete(id)
	})
	return
}

// Forward returns the id corresponding to the given label.
//
// If the label is not contained in this map, the zero ID is returned.
// The zero ID is never returned for a valid id.
func (mp *IMap[Label]) Forward(label Label) (ID, error) {
	return mp.forward.GetZero(label)
}

// Reverse returns the label corresponding to the given id.
// When id is not contained in this map, the zero value of the label type is contained.
func (mp *IMap[Label]) Reverse(id ID) (Label, error) {
	return mp.reverse.GetZero(id)
}

// IdentityMap writes canonical label mappings to the given storage.
//
// Concretely a pair (L1, L2) is written to storage iff
//
//	mp.Reverse(mp.Forward(L1)) == L2 && L1 != L2
func (mp *IMap[Label]) IdentityMap(storage Storage[Label, Label]) error {
	return mp.forward.Iterate(func(label Label, id ID) error {
		value, err := mp.reverse.GetZero(id)
		if err != nil {
			return err
		}
		if value != label {
			return storage.Set(label, value)
		}
		return nil
	})
}

// Close closes any storages related to this IMap.
//
// Calling close multiple times results in err = nil.
func (mp *IMap[Label]) Close() error {
	var errors [2]error

	if mp.forward != nil {
		errors[0] = mp.forward.Close()
		mp.forward = nil
	}
	if mp.reverse != nil {
		errors[1] = mp.reverse.Close()
		mp.reverse = nil
	}

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
