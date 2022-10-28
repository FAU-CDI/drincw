package imap

import "io"

// Engine creates Storages for an IMap
type Engine[Label comparable] interface {
	Forward() (Storage[Label, ID], error)
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

// MemoryEngine represents an engine that stores storages in memory
type MemoryEngine[Label comparable] struct{}

func (MemoryEngine[Label]) Forward() (Storage[Label, ID], error) {
	return make(MemoryStorage[Label, ID]), nil
}

func (MemoryEngine[Label]) Reverse() (Storage[ID, Label], error) {
	return make(MemoryStorage[ID, Label]), nil
}

// MemoryStorage implements Storage as an in-memory map
type MemoryStorage[Key comparable, Value any] map[Key]Value

func (ims MemoryStorage[Key, Value]) Set(key Key, value Value) error {
	ims[key] = value
	return nil
}

// Get returns the given value if it exists
func (ims MemoryStorage[Key, Value]) Get(key Key) (Value, bool, error) {
	value, ok := ims[key]
	return value, ok, nil
}

// GetZero returns the value associated with Key, or the zero value otherwise.
func (ims MemoryStorage[Key, Value]) GetZero(key Key) (Value, error) {
	return ims[key], nil
}

func (ims MemoryStorage[Key, Value]) Has(key Key) (bool, error) {
	_, ok := ims[key]
	return ok, nil
}

// Delete deletes the given key from this storage
func (ims MemoryStorage[Key, Value]) Delete(key Key) error {
	delete(ims, key)
	return nil
}

// Iterate calls f for all entries in Storage.
// there is no guarantee on order.
func (ims MemoryStorage[Key, Value]) Iterate(f func(Key, Value) error) error {
	for key, value := range ims {
		if err := f(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Close closes this MapStorage, deleting all values
func (ims MemoryStorage[Key, Value]) Close() error {
	for key := range ims {
		delete(ims, key)
	}
	return nil
}
