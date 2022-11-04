package imap

// MemoryEngine represents an engine that stores storages in memory
type MemoryEngine[Label comparable] struct{}

func (MemoryEngine[Label]) Forward() (Storage[Label, [2]ID], error) {
	ms := make(MemoryStorage[Label, [2]ID])
	return &ms, nil
}

func (MemoryEngine[Label]) Reverse() (Storage[ID, Label], error) {
	ms := make(MemoryStorage[ID, Label])
	return &ms, nil
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
func (ims *MemoryStorage[Key, Value]) Close() error {
	*ims = nil
	return nil
}
