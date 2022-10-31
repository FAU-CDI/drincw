// Package iterator provides Generic Iterator and Generator Interfaces.
package iterator

import "context"

// Iterator represents an object that can be iterated over.
// It is not safe for concurrent access.
//
// A user of an iterator must ensure that the iterator is closed.
// A user should furthermore use the Err() method to check if an error occured.
//
// A typical use of an iterator would be something like:
//
//	var it Iterator[T] // get the iterator from somewhere
//	defer it.Close()
//
//	for it.Next() {
//	  datum := it.Datum()
//	  _ = datum // ... do something with datum ...
//	}
//
//	if err := it.Err(); err != nil {
//	  return err // an error occured!
//	}
type Iterator[T any] interface {
	// Next advances this iterator to the next value.
	// The returned value indicates if there are further values.
	Next() bool

	// Datum returns the current value of this iterator.
	Datum() T

	// Err returns any error that occured during iteration.
	Err() error

	// Close closes this iterator, indicating to the sender that no more
	// values would be received.
	Close() error
}

// Generator is the opposite of Iterator, allowing values to be sent to a receiving iterator.
// Methods on a generator may not be called concurrently.
type Generator[T any] interface {
	// Yield yields an item to the receiving end.
	// The return value indicates if the corresponding Iterator requested cancellation.
	Yield(datum T) bool

	// YieldError yields an error to the receiving end.
	// The return value indiciates if the corresponding Iterator requested cancellation or error is not nil.
	YieldError(err error) bool

	// Returned indicates if the Return method was called.
	Returned() bool

	// Return closes this generator.
	//
	// Calling Return multiple times is an error.
	// Calls to Yield and YieldError after Return are also an error.
	Return()
}

// impl implements both [Iterator] and [Generator]
// Values should be created using newImpl
type impl[T any] struct {
	context context.Context
	cancel  context.CancelFunc

	messages chan message[T]
	returned bool

	datum T
	err   error
}

func newImpl[T any]() *impl[T] {
	context, cancel := context.WithCancel(context.Background())
	return &impl[T]{
		context:  context,
		cancel:   cancel,
		messages: make(chan message[T]),
	}
}

type message[T any] struct {
	datum T
	err   error
}

func (it *impl[T]) Next() (ok bool) {
	select {
	case <-it.context.Done():
		return false
	case message, mok := <-it.messages:
		if !mok {
			return false
		}
		it.err = message.err
		it.datum = message.datum
	}

	if it.err != nil {
		it.cancel()
		return false
	}

	return true
}

func (it *impl[T]) Datum() T {
	return it.datum
}

// Close closes the iterator
func (it *impl[T]) Close() error {
	it.cancel()
	return nil
}

// Err returns any error that occured.
// It may not be called
func (it *impl[T]) Err() error {
	return it.err
}

// sending end

func (it *impl[T]) Yield(datum T) (closed bool) {
	return it.send(message[T]{
		datum: datum,
		err:   nil,
	})
}

func (it *impl[T]) YieldError(err error) (closed bool) {
	if err == nil {
		return false
	}
	return it.send(message[T]{
		err: err,
	}) || err != nil
}

func (it *impl[T]) send(message message[T]) (closed bool) {
	select {
	case it.messages <- message:
		return false
	case <-it.context.Done():
		return true
	}
}

func (it *impl[T]) Returned() bool {
	return it.returned
}

func (it *impl[T]) Return() {
	close(it.messages)
	it.returned = true
}
