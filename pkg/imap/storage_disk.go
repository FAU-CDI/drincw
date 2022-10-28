package imap

import (
	"encoding/json"

	"github.com/akrylysov/pogreb"
)

// MakeDiskStorage creates a new disk-based storage stored at path.
func MakeDiskStorage[Key comparable, Value any](path string) (*DiskStorage[Key, Value], error) {
	db, err := pogreb.Open(path, &pogreb.Options{})
	if err != nil {
		return nil, err
	}
	storage := &DiskStorage[Key, Value]{
		DB: db,

		MarshalKey: func(key Key) ([]byte, error) {
			return json.Marshal(key)
		},
		UnmarshalKey: func(dest *Key, src []byte) error {
			return json.Unmarshal(src, dest)
		},
		MarshalValue: func(value Value) ([]byte, error) {
			return json.Marshal(value)
		},
		UnmarshalValue: func(dest *Value, src []byte) error {
			return json.Unmarshal(src, dest)
		},
	}
	return storage, nil
}

// DiskStorage implements Storage as an in-memory storage
type DiskStorage[Key comparable, Value any] struct {
	DB *pogreb.DB

	MarshalKey     func(key Key) ([]byte, error)
	UnmarshalKey   func(dest *Key, src []byte) error
	MarshalValue   func(value Value) ([]byte, error)
	UnmarshalValue func(dest *Value, src []byte) error
}

func (ds *DiskStorage[Key, Value]) Set(key Key, value Value) error {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		return err
	}
	valueB, err := ds.MarshalValue(value)
	if err != nil {
		return err
	}

	return ds.DB.Put(keyB, valueB)
}

// Get returns the given value if it exists
func (ds *DiskStorage[Key, Value]) Get(key Key) (v Value, b bool, err error) {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		return v, b, err
	}

	// check if we have the key
	ok, err := ds.DB.Has(keyB)
	if err != nil {
		return v, b, err
	}
	if !ok {
		return v, false, nil
	}

	valueB, err := ds.DB.Get(keyB)
	if err != nil {
		return v, b, err
	}

	if err := ds.UnmarshalValue(&v, valueB); err != nil {
		return v, b, err
	}

	return v, true, nil
}

// GetZero returns the value associated with Key, or the zero value otherwise.
func (ds *DiskStorage[Key, Value]) GetZero(key Key) (Value, error) {
	value, _, err := ds.Get(key)
	return value, err
}

func (ds *DiskStorage[Key, Value]) Has(key Key) (bool, error) {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		return false, err
	}

	ok, err := ds.DB.Has(keyB)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Delete deletes the given key from this storage
func (ds *DiskStorage[Key, Value]) Delete(key Key) error {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		return err
	}

	if err := ds.DB.Delete(keyB); err != nil {
		return err
	}

	return nil
}

// Iterate calls f for all entries in Storage.
// there is no guarantee on order.
func (ds *DiskStorage[Key, Value]) Iterate(f func(Key, Value) error) error {
	it := ds.DB.Items()
	for {
		keyB, valueB, err := it.Next()
		if err == pogreb.ErrIterationDone {
			break
		}
		if err != nil {
			return err
		}

		var key Key
		if err := ds.UnmarshalKey(&key, keyB); err != nil {
			return err
		}
		var value Value
		if err := ds.UnmarshalValue(&value, valueB); err != nil {
			return err
		}
		if err := f(key, value); err != nil {
			return err
		}
	}
	return nil
}
