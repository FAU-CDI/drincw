package imap

import "io"

// Storage represents a storage for an imap instance.
//
// Must be able to handle multiple reading operations concurrently.
type Storage[Key comparable, Value any] interface {
	io.Closer

	// Set sets the given key to the given value
	Set(key Key, value Value)

	// Get retrieves the value for Key from the given storage.
	// The second value indiciates if the value was found.
	Get(key Key) (Value, bool)

	// GetZero is like Get, but when the value does not exist returns the zero value
	GetZero(key Key) Value

	// Has is like Get, but returns only the second value.
	Has(key Key) bool

	// Delete deletes the given key from this storage
	Delete(key Key)

	// Iterate calls f for all entries in Storage.
	// there is no guarantee on order.
	Iterate(f func(Key, Value))
}

// MapStorage implements Storage as an in-memory map
type MapStorage[Key comparable, Value any] map[Key]Value

func (ims MapStorage[Key, Value]) Set(key Key, value Value) {
	ims[key] = value
}

// Get returns the given value if it exists
func (ims MapStorage[Key, Value]) Get(key Key) (Value, bool) {
	value, ok := ims[key]
	return value, ok
}

// GetZero returns the value associated with Key, or the zero value otherwise.
func (ims MapStorage[Key, Value]) GetZero(key Key) Value {
	return ims[key]
}

func (ims MapStorage[Key, Value]) Has(key Key) bool {
	_, ok := ims[key]
	return ok
}

// Delete deletes the given key from this storage
func (ims MapStorage[Key, Value]) Delete(key Key) {
	delete(ims, key)
}

// Iterate calls f for all entries in Storage.
// there is no guarantee on order.
func (ims MapStorage[Key, Value]) Iterate(f func(Key, Value)) {
	for key, value := range ims {
		f(key, value)
	}
}

// Close closes this MapStorage, deleting all values
func (ims MapStorage[Key, Value]) Close() error {
	for key := range ims {
		delete(ims, key)
	}
	return nil
}
