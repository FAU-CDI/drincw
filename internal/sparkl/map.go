package sparkl

type LabelMap[Label comparable] struct {
	forward map[Label]indexID
	reverse map[indexID]Label

	id indexID // last id inserted
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
