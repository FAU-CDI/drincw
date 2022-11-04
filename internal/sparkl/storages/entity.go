package storages

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
)

// sEntityPool is a pool of stored entities
var sEntityPool = sync.Pool{
	New: func() any {
		return new(sEntity)
	},
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// sEntity represents a stored entity that does not hold references to child entitites
type sEntity struct {
	URI  wisski.URI   // URI of this entity
	Path []wisski.URI // the path of this entity

	Fields   map[string][]wisski.FieldValue // values for specific fields
	Children map[string][]wisski.URI        // child entities
}

// Reset resets this stored entity
func (s *sEntity) Reset() {
	s.Path = nil
	s.Children = nil
	s.Fields = nil
	s.URI = ""
}

// Encode encodes this stored entity into a stream of bytes
func (s *sEntity) Encode() ([]byte, error) {
	// take a buffer
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()

	// encode the entity
	err := gob.NewEncoder(buffer).Encode(s)
	if err != nil {
		return nil, err
	}

	// return a copy of the buffer!
	bytes := buffer.Bytes()
	data := make([]byte, len(bytes))
	copy(data, bytes)

	return data, nil
}

// Decode decodes this stored entity from a stream of bytes
func (s *sEntity) Decode(data []byte) error {
	// take a buffer
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	// fill it with data
	buffer.Reset()
	buffer.Write(data)

	// and decode from it
	return gob.NewDecoder(buffer).Decode(s)
}

func init() {
	gob.Register(wisski.URI(""))
	gob.Register(wisski.FieldValue{})
	gob.Register(sEntity{})
}