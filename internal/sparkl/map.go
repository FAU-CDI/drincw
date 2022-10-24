package sparkl

type LabelMap[Label comparable] struct {
	forward map[Label]indexID
	reverse map[indexID]Label

	id indexID // last id inserted
}

// Identify marks the left and right label as being identical labels
//
// Any object currently carrying the same id as the left element will now carry the id of the right.
// Future calls to reverse will never return right.
func (mp *LabelMap[Label]) Identify(left, right Label) {
	// left and right are the same object
	canonical := mp.Add(left)
	alias := mp.Add(right)

	// already in the same group => nothing to do
	if canonical == alias {
		return
	}

	for label, id := range mp.forward {
		if id != alias || label == left {
			continue
		}
		mp.forward[label] = canonical

		// delete the reverse mapping of the alias
		// because it cannot ever be returned
		delete(mp.reverse, id)
	}
}

func (mp *LabelMap[Label]) IdentityMap() map[Label]Label {
	canonmap := make(map[Label]Label)

	for label, id := range mp.forward {
		if mp.reverse[id] != label {
			canonmap[label] = mp.reverse[id]
		}
	}

	return canonmap
}

func (mp *LabelMap[Label]) Reset() {
	mp.forward = make(map[Label]indexID)
	mp.reverse = make(map[indexID]Label)
	mp.id = indexID(0)
}

// Next returns a fresh ID from the LabelMap
func (mp *LabelMap[Label]) Next() indexID {
	return mp.id.next()
}

func (mp *LabelMap[Label]) Add(label Label) indexID {
	if index, ok := mp.forward[label]; ok {
		return index
	}
	id := mp.id.next()
	mp.forward[label] = id
	mp.reverse[id] = label
	return id
}

func (mp *LabelMap[Label]) Forward(label Label) indexID {
	return mp.forward[label]
}

func (mp *LabelMap[Label]) Reverse(id indexID) Label {
	return mp.reverse[id]
}

// indexID represents an item in the index
type indexID int64

// next increments this ID, and then returns a copy of the new value.
// It is the equivalent of the "++" operator.
func (i *indexID) next() indexID {
	(*i)++
	return *i
}
