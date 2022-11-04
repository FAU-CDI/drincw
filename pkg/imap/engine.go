package imap

import "io"

// Engine creates Storages for an IMap
type Engine[Label comparable] interface {
	Forward() (Storage[Label, [2]ID], error)
	Reverse() (Storage[ID, Label], error)
}

// Storage represents a storage for an imap instance.
//
// Must be able to handle multiple reading operations concurrently.
type Storage[Key comparable, Value any] interface {
	io.Closer

	// Set sets the given key to the given value
	Set(key Key, value Value) error

	// Get retrieves the value for Key from the given storage.
	// The second value indiciates if the value was found.
	Get(key Key) (Value, bool, error)

	// GetZero is like Get, but when the value does not exist returns the zero value
	GetZero(key Key) (Value, error)

	// Has is like Get, but returns only the second value.
	Has(key Key) (bool, error)

	// Delete deletes the given key from this storage
	Delete(key Key) error

	// Iterate calls f for all entries in Storage.
	//
	// When any f returns a non-nil error, that error is returned immediatly to the caller
	// and iteration stops.
	//
	// There is no guarantee on order.
	Iterate(f func(Key, Value) error) error
}
