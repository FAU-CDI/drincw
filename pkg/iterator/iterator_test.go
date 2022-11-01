// Package iterator provides Generic Iterator and Generator Interfaces.
package iterator

import (
	"errors"
	"fmt"
)

func ExampleNew() {
	iterator := New(func(generator Generator[string]) {
		if generator.Yield("hello") {
			return
		}
		if generator.Yield("world") {
			return
		}
	})
	defer iterator.Close()

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())

	// Output: hello
	// world
	// <nil>
}

func ExampleNew_close() {
	iterator := New(func(generator Generator[string]) {
		if generator.Yield("hello") {
			return
		}

		if generator.Yield("world") {
			return
		}

		panic("never reached")
	})
	defer iterator.Close()

	for iterator.Next() {
		fmt.Println(iterator.Datum())
		iterator.Close()
	}

	fmt.Println(iterator.Err())

	// Output: hello
	// <nil>
}

func ExampleNew_error() {
	iterator := New(func(generator Generator[string]) {
		if generator.Yield("hello") {
			return
		}

		generator.YieldError(errors.New("something went wrong"))
	})
	defer iterator.Close()

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}

	fmt.Println(iterator.Err())

	// Output: hello
	// something went wrong
}
