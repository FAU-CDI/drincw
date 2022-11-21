package iterator

// New creates a new iterator generator pair and returns the iterator.
//
// The generator is passed to the function source.
// Once source returns, the return method on the generator is called if it has not been already.
func New[T any](source func(generator Generator[T])) Iterator[T] {
	it := newImpl[T]()
	go func(it Generator[T]) {
		source(it)
		if !it.Returned() {
			it.Return()
		}
	}(it)
	return it
}

// Slice creates a new Iterator that yields elements from the given slice
func Slice[T any](elements []T) Iterator[T] {
	return New(func(sender Generator[T]) {
		defer sender.Return()

		for _, element := range elements {
			if sender.Yield(element) {
				break
			}
		}
	})
}

// Map creates a new iterator that produces the same values as source, but mapped over f.
// If source produces an error, the returned iterator also produces the same error.
func Map[Element1, Element2 any](source Iterator[Element1], f func(Element1) Element2) Iterator[Element2] {
	return New(func(sender Generator[Element2]) {
		defer sender.Return()

		for source.Next() {
			sender.Yield(f(source.Datum()))
		}
		sender.YieldError(source.Err())
	})
}

// Connect creates a new iterator that calls f for every element returned by source.
// If the pipe function returns true, iteration over the original elements stops.
func Connect[Element1, Element2 any](source Iterator[Element1], f func(element Element1, sender Generator[Element2]) (closed bool)) Iterator[Element2] {
	return New(func(sender Generator[Element2]) {
		// close the source
		defer source.Close()

		// close the sender unless we already have
		defer func() {
			if sender.Returned() {
				return
			}
			if err := source.Err(); err != nil {
				sender.YieldError(err)
			}
		}()

		for source.Next() {
			if f(source.Datum(), sender) {
				break
			}
			if sender.Returned() {
				break
			}
		}
	})
}

// Pipe pipes elements from src into dst.
// If any error occurs in src, the same error is sent to dst.
//
// The boolean indicates if the caller should continue running.
func Pipe[Element any](dst Generator[Element], src Iterator[Element]) bool {
	for src.Next() {
		if dst.Yield(src.Datum()) {
			return true
		}
	}
	return dst.YieldError(src.Err())
}

// Drain iterates all values in it until no more values are returned.
// All returned values are stored in a slice which is returned to the caller.
func Drain[Element any](it Iterator[Element]) ([]Element, error) {
	defer it.Close()

	var drain []Element
	for it.Next() {
		drain = append(drain, it.Datum())
	}
	return drain, it.Err()
}
