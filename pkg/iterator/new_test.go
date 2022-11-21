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

func ExampleSlice() {
	iterator := Slice([]int{1, 2, 3, 4, 5})
	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExampleMap() {
	iterator :=
		Map(
			Slice([]int{1, 2, 3, 4, 5}),
			func(value int) bool {
				return value%2 == 0
			},
		)
	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: false
	// true
	// false
	// true
	// false
	// <nil>
}

func ExampleConnect() {
	source := Slice([]int{0, 1, 2})
	dest := Connect(source, func(element int, sender Generator[int]) (closed bool) {
		sender.Yield(2 * element)
		sender.Yield(2*element + 1)
		return false
	})

	for dest.Next() {
		fmt.Println(dest.Datum())
	}
	fmt.Println(dest.Err())
	// Output: 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExamplePipe() {
	iterator := New(func(generator Generator[int]) {
		source := Slice([]int{0, 1, 2, 3, 4, 5})
		Pipe(generator, source)
	})

	for iterator.Next() {
		fmt.Println(iterator.Datum())
	}
	fmt.Println(iterator.Err())
	// Output: 0
	// 1
	// 2
	// 3
	// 4
	// 5
	// <nil>
}

func ExampleDrain() {
	source := Slice([]int{0, 1, 2, 3, 4, 5})
	fmt.Println(Drain(source))
	// Output: [0 1 2 3 4 5] <nil>
}
