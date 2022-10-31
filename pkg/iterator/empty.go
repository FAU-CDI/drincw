package iterator

// Empty returns a new iterator that contains no elements.
// The Err method will return err, which may be nil.
func Empty[Element any](err error) Iterator[Element] {
	return empty[Element]{err: err}
}

type empty[Element any] struct {
	err error
}

func (empty[Element]) Next() bool {
	return false
}

func (empty[Element]) Datum() (element Element) {
	return
}
func (empty empty[Element]) Err() error {
	return empty.err
}

func (err empty[Element]) Close() error {
	return nil
}
