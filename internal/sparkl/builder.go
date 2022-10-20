package sparkl

type ArrayBuilder[T any] struct {
	array []T
}

func NewArrayBuilder[T any](len int) *ArrayBuilder[T] {
	return &ArrayBuilder[T]{
		array: make([]T, len),
	}
}

func (ab *ArrayBuilder[T]) Len() int {
	return len(ab.array)
}

func (ab *ArrayBuilder[T]) Get(i int) *T {
	return &ab.array[i]
}

func (ab *ArrayBuilder[T]) Grow(count int) {
	// if there isn't enough space in the buffer
	// grow it by the required amount!
	if cap(ab.array)-len(ab.array) < count {
		old := ab.array

		// make a new slice (with enough space)
		// and copy over the array
		ab.array = make([]T, len(old)+count)
		copy(ab.array, old)

		// zero out the old array
		var zero T
		for i := range old {
			old[i] = zero
		}
	}
}

func (ab *ArrayBuilder[T]) Build() []T {
	return ab.array
}

func (ab *ArrayBuilder[T]) Reset(len int) {
	ab.array = make([]T, 0)
}
