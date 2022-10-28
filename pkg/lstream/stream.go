package lstream

// Stream represents a stream that is evaluated lazily
type Stream[Element any] interface {
	// Next returns the next element in this stream.
	//
	// ok indiciates if there are more elements left in this channel.
	// error indiciates if an error occured
	Next() (element Element, ok bool, err error)
}

// New creates a new stream
func New[Element any](source func(sender chan<- Element) error) Stream[Element] {
	c := make(chan Element)

	stream := lstream[Element]{
		c: c,
	}

	go func() {
		defer close(c)

		stream.err = source(c)
	}()

	return &stream
}

// NewConcrete creates a new stream from a given set of elements
func NewConcrete[Element any](elements []Element) Stream[Element] {
	return &cstream[Element]{
		elements: elements,
		index:    0,
	}
}

// Pipe calls pipe for every element of the stream, and creates a new result stream
func Pipe[Element any](s Stream[Element], pipe func(element Element, sender chan<- Element) error) Stream[Element] {
	return New(func(sender chan<- Element) (err error) {
		c := Channel(s, &err)
		for element := range c {
			if err := pipe(element, sender); err != nil {
				// make sure that the channel is drained
				for range c {
				}
				return err
			}
		}
		return nil
	})
}

// Drain drains the entire stream into a slice
func Drain[Element any](s Stream[Element]) ([]Element, error) {
	var drain []Element
	var err error
	for element := range Channel(s, &err) {
		drain = append(drain, element)
	}
	return drain, err
}

// Channel returns a channel representing the underlying stream.
// The channel will receive values as long as there are values.
// The channel must be drained by the caller.
//
// If an error occurs, writes the error to errDst before closing the channel
func Channel[Element any](str Stream[Element], errDst *error) <-chan Element {
	c := make(chan Element)
	go func() {
		defer close(c)
		for {
			elem, ok, err := str.Next()
			if !ok || err != nil {
				*errDst = err
				break
			}
			c <- elem
		}
	}()
	return c
}

type lstream[Element any] struct {
	c chan Element

	// any error that occured; should only be read once c is closed
	err error
}

func (*lstream[Element]) isStream() {}

func (l *lstream[Element]) Next() (Element, bool, error) {
	element, ok := <-l.c
	if !ok {
		return element, false, l.err
	}
	return element, ok, nil
}

type cstream[Element any] struct {
	elements []Element
	index    int
}

func (*cstream[Element]) isStream() {}

func (c *cstream[Element]) Next() (element Element, ok bool, err error) {
	// no more elements left, so we can clear everything!
	if len(c.elements) <= c.index {
		c.elements = nil
		c.index = 0
		return
	}

	element = c.elements[c.index]
	c.index++
	return element, true, nil
}
