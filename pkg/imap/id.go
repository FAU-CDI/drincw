package imap

import (
	"errors"
	"fmt"
	"math/big"
)

// ID represents an ID of a specific element.
// Not all IDs are valid, see [Valid].
//
// Users should not rely on the exact size of this data type.
// Instead they should use appropriate methods to compare values.
//
// Internally, an ID is represented in big endian array of bytes.
// It effectively corresponds to a uint32.
type ID [4]byte

// IDLen is the length of the ID type
const IDLen = len(ID{})

// Valid checks if this ID is valid
func (id ID) Valid() bool {
	for i := IDLen - 1; i >= 0; i-- {
		// we start this loop at the back
		// because most likely the id is going to be at the small end
		if id[i] != 0 {
			return true
		}
	}
	return false
}

// Reset resets this id to an invalid value
func (id *ID) Reset() {
	// TODO: Untested
	for i := 0; i < IDLen; i++ {
		(*id)[i] = 0
	}
}

// Inc increments this ID, returning a copy of the new value.
// It is the equivalent of the "++" operator.
//
// When Inc() exceeds the maximum possible value for an ID, panics.
func (id *ID) Inc() ID {
	for i := IDLen - 1; i >= 0; i-- {
		(*id)[i]++
		if (*id)[i] != 0 {
			return *id
		}
	}

	// NOTE(twiesing): If this line is ever reached we should increase the size of the ID type.
	panic("Inc: Overflow")
}

// Int writes the numerical value of this id into the given big int.
// The big.Int is returned for convenience.
func (id ID) Int(value *big.Int) *big.Int {
	bytes := make([]byte, IDLen)
	id.MarshalTo(bytes)
	return value.SetBytes(bytes)
}

// LoadInt sets the value of this id as an integer and returns it.
// Trying to load an integer bigger than the maximal id results in a panic.
//
// The ID is returned for convenience.
func (id *ID) LoadInt(value *big.Int) *ID {
	id.UnmarshalFrom(value.FillBytes(make([]byte, IDLen)))
	return id
}

// Less compares this ID to another id.
// An id is less than another id iff Inc() has been called fewer times.
func (id ID) Less(other ID) bool {
	for i := 0; i < IDLen; i++ {
		if id[i] < other[i] {
			return true
		}
		if id[i] > other[i] {
			return false
		}
	}
	return false
}

// String formats this id as a string.
// It is only intended for debugging.
func (id ID) String() string {
	return fmt.Sprintf("%v", [IDLen]byte(id))
}

// MarshalTo encodes id using a big endian encoding into dest.
// dest must be of at least size [IDLen]
func (id ID) MarshalTo(dest []byte) {
	_ = dest[IDLen-1] // boundary hint to compiler
	for i := 0; i < IDLen; i++ {
		dest[i] = id[i]
	}
}

// UnmarshalFrom encodes an id using big endian encoding from src.
// src must be of at least size [IDLen]
func (id *ID) UnmarshalFrom(src []byte) {
	_ = src[IDLen-1] // boundary hint to compiler
	for i := 0; i < IDLen; i++ {
		(*id)[i] = src[i]
	}
}

// TODO: These require testing

// EncodeIDs marshals a set of ids into a new byte slice.
func EncodeIDs(ids ...ID) []byte {
	bytes := make([]byte, len(ids)*IDLen)
	for i := 0; i < len(ids); i++ {
		ids[i].MarshalTo(bytes[i*IDLen:])
	}
	return bytes
}

// DecodeIDs unmarshals a set of ids encoded with MarshalIDs.
// When src is not an integer multiple of IDLen, any trailing data is ignored.
func DecodeIDs(src []byte) []ID {
	ids := make([]ID, len(src)/IDLen)
	for i := 0; i < len(ids); i++ {
		ids[i].UnmarshalFrom(src[i*IDLen:])
	}
	return ids
}

// DecodeID works exactly like DecodeIDs, but unmarshals only the id
// with the given index.
//
// If src does not contain enough bytes for the given index, the behavior is undefined.
func DecodeID(src []byte, index int) (id ID) {
	id.UnmarshalFrom(src[index*IDLen:])
	return
}

// MarshalID is like value.MarshalTo, but automatically creates a new slice
// and returns nil error.
func MarshalID(value ID) ([]byte, error) {
	dest := make([]byte, IDLen)
	value.MarshalTo(dest)
	return dest, nil
}

var errUnmarshal = errors.New("unmarshalID: invalid length")

// UnmarshalID is like dest.UnmarshalFrom,
// but returns an error when the given slice is not of the correct length.
func UnmarshalID(dest *ID, src []byte) error {
	if len(src) < IDLen {
		return errUnmarshal
	}
	dest.UnmarshalFrom(src)
	return nil
}
