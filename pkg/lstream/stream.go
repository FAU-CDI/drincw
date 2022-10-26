package lstream

// Stream represents a stream that is evaluated lazily
type Stream[Element any] interface {
	// Next returns the next element in this stream
	Next() (element Element, ok bool)
}

// New creates a new stream
func New[Element any](source func(sender chan<- Element)) Stream[Element] {
	c := make(chan Element)

	go func() {
		defer close(c)
		source(c)
	}()

	return &lstream[Element]{
		c: c,
	}
}

// NewConcrete creates a new stream from a given set of elements
func NewConcrete[Element any](elements []Element) Stream[Element] {
	return &cstream[Element]{
		elements: elements,
		index:    0,
	}
}

// Pipe calls pipe for every element of the stream, and creates a new result stream
func Pipe[Element any](s Stream[Element], pipe func(element Element, sender chan<- Element)) Stream[Element] {
	return New(func(sender chan<- Element) {
		for element := range Channel(s) {
			pipe(element, sender)
		}
	})
}

// Drain drains the entire stream into a slice
func Drain[Element any](s Stream[Element]) []Element {
	var drain []Element
	for element := range Channel(s) {
		drain = append(drain, element)
	}
	return drain
}

// Channel represents a stream as a channel
func Channel[Element any](str Stream[Element]) <-chan Element {
	if s, ok := str.(*lstream[Element]); ok {
		return s.c
	}

	c := make(chan Element)
	go func() {
		defer close(c)
		for {
			elem, ok := str.Next()
			if !ok {
				break
			}
			c <- elem
		}
	}()
	return c
}

type lstream[Element any] struct {
	c chan Element
}

func (*lstream[Element]) isStream() {}

func (l *lstream[Element]) Next() (Element, bool) {
	element, ok := <-l.c
	return element, ok
}

type cstream[Element any] struct {
	elements []Element
	index    int
}

func (*cstream[Element]) isStream() {}

func (c *cstream[Element]) Next() (element Element, ok bool) {
	// no more elements left, so we can clear everything!
	if len(c.elements) <= c.index {
		c.elements = nil
		c.index = 0
		return
	}

	element = c.elements[c.index]
	c.index++
	return element, true
}
