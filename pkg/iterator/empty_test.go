package iterator

import (
	"errors"
	"fmt"
)

func ExampleEmpty() {
	thing := Empty[any](nil)
	fmt.Println(thing.Next())
	fmt.Println(thing.Err())

	// Output: false
	// <nil>
}

func ExampleEmpty_error() {
	thing := Empty[any](errors.New("some error"))
	fmt.Println(thing.Next())
	fmt.Println(thing.Err())

	// Output: false
	// some error
}
