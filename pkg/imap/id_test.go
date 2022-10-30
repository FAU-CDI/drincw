package imap

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
)

func ExampleID() {

	// create a new id -- which isn't valid
	var id ID
	fmt.Println(id)
	fmt.Println(id.Valid())

	// increment the id -- it is now valid
	fmt.Println(id.Inc())
	fmt.Println(id.Valid())

	// create the value 10
	var ten ID
	for i := 0; i < 10; i++ {
		ten.Inc()
	}

	// compare it to some other id -- it is no longer valid
	fmt.Println(id.Less(ten))

	// Output: [0 0 0 0 0 0 0 0]
	// false
	// [0 0 0 0 0 0 0 1]
	// true
	// true
}

// maximum numbers for the ID "torture tests"
const (
	largeIDMax = (1 << (3 * 8))       // use a full 3 bytes
	smallIDMax = (1 << (8 + (8 / 2))) // use 2.5 bytes
)

func TestID_Int(t *testing.T) {
	var (
		id ID      // ID representation
		bi big.Int // Big Integer representation
	)

	// increment, which is guaranteed to have that value
	for i := 0; i < largeIDMax; i++ {
		bi.SetInt64(-1) // store a dirty value into the integer
		id.Int(&bi)     // decode the value

		value := int(bi.Int64())
		if value != i {
			t.Error("failed to decode incrementally", i)
		}
		id.Inc()
	}

	// next encode and decode again
	// then check if the values are identical
	for i := 0; i < largeIDMax; i++ {
		id.LoadInt(bi.SetInt64(int64(i))) // store the integer into the id

		bi.SetInt64(-1) // set a dirty value in the bigint
		id.Int(&bi)     // decode the id again

		got := int(bi.Int64())
		if got != i {
			t.Error("failed to round trip", i)
		}
	}
}

func BenchmarkID_Inc(b *testing.B) {
	var id ID
	for i := 0; i < b.N; i++ {
		id.Reset()
		for j := 0; j < smallIDMax; j++ {
			id.Inc()
		}
	}
}

// Test that only non-zero values are valid
func TestID_Valid(t *testing.T) {
	var (
		id ID      // ID representation
		bi big.Int // to load big integers
	)

	for i := 0; i < largeIDMax; i++ {
		id.LoadInt(bi.SetInt64(int64(i)))

		got := id.Valid()
		want := i != 0

		if got != want {
			t.Errorf("ID(%d).Valid() = %v, want = %v", i, got, want)
		}
	}
}

func BenchmarkID_Less(b *testing.B) {
	var idI, idJ ID

	idI.LoadInt(big.NewInt(10000))
	idJ.LoadInt(big.NewInt(12))

	for i := 0; i < b.N; i++ {
		idI.Less(idJ)
	}
}

// Test that the order of ids behaves as expected.
func TestID_Order(t *testing.T) {
	var (
		idI, idJ ID
		big      big.Int
	)

	bytesI := make([]byte, IDLen)
	bytesJ := make([]byte, IDLen)

	// check that the .Less() method indeed implements the order
	// that was induced by their generation
	for i := 0; i < smallIDMax; i++ {
		idI.LoadInt(big.SetInt64(int64(i))) // set i to the right value
		idI.MarshalTo(bytesI)               // and decode the bytes

		for j := 0; j < smallIDMax; j++ {
			idJ.LoadInt(big.SetInt64(int64(j))) // set j
			idJ.MarshalTo(bytesJ)               // and decode the bytes

			{
				got := idI.Less(idJ)
				want := i < j
				if got != want {
					t.Errorf("id(%d) < id(%d) = %v, want %v", i, j, got, want)
				}
			}

			{
				got := idJ.Less(idI)
				want := j < i
				if got != want {
					t.Errorf("id(%d) < id(%d) = %v, want %v", j, i, got, want)
				}
			}

			{
				got := bytes.Compare(bytesI, bytesJ)
				want := 0
				if i < j {
					want = -1
				} else if i > j {
					want = 1
				}
				if got != want {
					t.Errorf("compare(id(%d), id(%d)) = %v, want %v", j, i, got, want)
				}
			}
		}
	}
}
