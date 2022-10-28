package imap

import (
	"encoding/binary"
	"errors"
	"strconv"
)

// ID represents an ID of a specific element.
//
// NOTE(twiesing): This is currently of size 1.
// It may increase in the future.
type ID [1]uint64

// Valid checks if this id is valid, meaning it is not the zero ID.
func (id ID) Valid() bool {
	return uint64(id[0]) != 0
}

// Reset resets this id to an invalid value.
func (id *ID) Reset() {
	id[0] = 0
}

func (id ID) String() string {
	return strconv.FormatUint(uint64(id[0]), 10)
}

// Inc increments this ID, and then returns a copy of the new value.
// It is the equivalent of the "++" operator.
func (id *ID) Inc() ID {
	(*id)[0]++
	return *id
}

// Less returns true if this ID is less than the provided other ID
func (id ID) Less(other ID) bool {
	return uint64(id[0]) < uint64(other[0])
}

// MarshalID encodes an id as a slice of bytes
func MarshalID(id ID) ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id[0])
	return b, nil
}

var errWrongLength = errors.New("UnmarshalID: Wrong length")

// UnmarshalID decodes an id from a slice of bytes
func UnmarshalID(id *ID, data []byte) error {
	if len(data) != 8 {
		return errWrongLength
	}
	(*id)[0] = binary.LittleEndian.Uint64(data)
	return nil
}
