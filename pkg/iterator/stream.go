package iterator

// New creates a new iterator generator pair and returns the iterator.
//
// The generator is passed to the function source.
// Once source returns, the return method on the generator is called if it has not been already.
func New[Element any](source func(generator Generator[Element])) Iterator[Element] {
	it := newImpl[Element]()
	go func(it Generator[Element]) {
		source(it)
		if !it.Returned() {
			it.Return()
		}
	}(it)
	return it
}

// NewFromElements creates a new Iterator that yields the given elements.
func NewFromElements[Element any](elements []Element) Iterator[Element] {
	return New(func(sender Generator[Element]) {
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

// Pipe creates a new iterator that calls pipe for every element returend by source.
// If the pipe function returns true, iteration over the original elements stops.
func Pipe[Element1, Element2 any](source Iterator[Element1], pipe func(element Element1, sender Generator[Element2]) (closed bool)) Iterator[Element2] {
	return New(func(sender Generator[Element2]) {
		// close the source
		defer source.Close()

		// close the sender if we already have
		defer func() {
			if sender.Returned() {
				return
			}
			if err := source.Err(); err != nil {
				sender.YieldError(err)
			}
		}()

		for source.Next() {
			if pipe(source.Datum(), sender) {
				break
			}
			if sender.Returned() {
				break
			}
		}
	})
}

// Drain iterates over values in it until no more values are returned.
// All returned values are stored in a slice which is returned to the user.
func Drain[Element any](it Iterator[Element]) ([]Element, error) {
	defer it.Close()

	var drain []Element
	for it.Next() {
		drain = append(drain, it.Datum())
	}
	return drain, it.Err()
}
