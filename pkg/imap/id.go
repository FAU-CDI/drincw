package imap

import "strconv"

// ID represents an ID of a specific element.
//
// It is always comparable, however the exact backing data type may change without notice.
type ID uint64

// Valid checks if this id is valid, meaning it is not the zero ID.
func (id ID) Valid() bool {
	return uint64(id) != 0
}

func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// Inc increments this ID, and then returns a copy of the new value.
// It is the equivalent of the "++" operator.
func (id *ID) Inc() ID {
	(*id)++
	return *id
}

// Less returns true if this ID is less than the provided other ID
func (id ID) Less(other ID) bool {
	return uint64(id) < uint64(other)
}
