package imap

import (
	"encoding/json"
	"log"

	"github.com/akrylysov/pogreb"
)

// MakeDiskStorage creates a new disk-based storage stored at path.
func MakeDiskStorage[Key comparable, Value any](path string) *DiskStorage[Key, Value] {
	db, err := pogreb.Open(path, &pogreb.Options{})
	if err != nil {
		log.Fatal(err)
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
	return storage
}

// DiskStorage implements Storage as an in-memory storage
type DiskStorage[Key comparable, Value any] struct {
	DB *pogreb.DB

	MarshalKey     func(key Key) ([]byte, error)
	UnmarshalKey   func(dest *Key, src []byte) error
	MarshalValue   func(value Value) ([]byte, error)
	UnmarshalValue func(dest *Value, src []byte) error
}

func (ds *DiskStorage[Key, Value]) Set(key Key, value Value) {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		log.Fatal(err)
	}
	valueB, err := ds.MarshalValue(value)
	if err != nil {
		log.Fatal(err)
	}

	ds.DB.Put(keyB, valueB)
}

// Get returns the given value if it exists
func (ds *DiskStorage[Key, Value]) Get(key Key) (v Value, b bool) {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		log.Fatal(err)
	}

	// check if we have the key
	ok, err := ds.DB.Has(keyB)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		return v, false
	}

	valueB, err := ds.DB.Get(keyB)
	if err != nil {
		log.Fatal(err)
	}

	if err := ds.UnmarshalValue(&v, valueB); err != nil {
		log.Fatal(err)
	}

	return v, true
}

// GetZero returns the value associated with Key, or the zero value otherwise.
func (ds *DiskStorage[Key, Value]) GetZero(key Key) Value {
	value, _ := ds.Get(key)
	return value
}

func (ds *DiskStorage[Key, Value]) Has(key Key) bool {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := ds.DB.Has(keyB)
	if err != nil {
		log.Fatal(err)
	}
	return ok
}

// Delete deletes the given key from this storage
func (ds *DiskStorage[Key, Value]) Delete(key Key) {
	keyB, err := ds.MarshalKey(key)
	if err != nil {
		log.Fatal(err)
	}

	if err := ds.DB.Delete(keyB); err != nil {
		log.Fatal(err)
	}
}

// Iterate calls f for all entries in Storage.
// there is no guarantee on order.
func (ds *DiskStorage[Key, Value]) Iterate(f func(Key, Value)) {
	it := ds.DB.Items()
	for {
		keyB, valueB, err := it.Next()
		if err == pogreb.ErrIterationDone {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var key Key
		if err := ds.UnmarshalKey(&key, keyB); err != nil {
			log.Fatal(err)
		}
		var value Value
		if err := ds.UnmarshalValue(&value, valueB); err != nil {
			log.Fatal(err)
		}
		f(key, value)
	}
}
