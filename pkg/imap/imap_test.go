package imap

import (
	"fmt"
	"strconv"
	"testing"
)

func ExampleIMap() {

	var mp IMap[string]
	mp.Reset()

	fmt.Println("add", mp.Add("hello"))
	fmt.Println("add", mp.Add("world"))
	fmt.Println("add", mp.Add("earth"))

	fmt.Println("add<again>", mp.Add("hello"))
	fmt.Println("add<again>", mp.Add("world"))
	fmt.Println("add<again>", mp.Add("earth"))

	fmt.Println("get", mp.Forward("hello"))
	fmt.Println("get", mp.Forward("world"))
	fmt.Println("get", mp.Forward("earth"))

	fmt.Println("reverse", mp.Reverse(1))
	fmt.Println("reverse", mp.Reverse(2))
	fmt.Println("reverse", mp.Reverse(3))

	mp.MarkIdentical("earth", "world")

	fmt.Println("reverse<again>", mp.Reverse(1))
	fmt.Println("reverse<again>", mp.Reverse(3))

	fmt.Println("add<again>", mp.Add("hello"))
	fmt.Println("add<again>", mp.Add("world"))
	fmt.Println("add<again>", mp.Add("earth"))

	// Output: add 1
	// add 2
	// add 3
	// add<again> 1
	// add<again> 2
	// add<again> 3
	// get 1
	// get 2
	// get 3
	// reverse hello
	// reverse world
	// reverse earth
	// reverse<again> hello
	// reverse<again> earth
	// add<again> 1
	// add<again> 3
	// add<again> 3
}

var testIDs [10000]string

func init() {
	for i := 0; i < 10000; i++ {
		testIDs[i] = strconv.Itoa(i)
	}
}

func BenchmarkIMap(b *testing.B) {
	var mp IMap[string]
	for i := 0; i < b.N; i++ {
		mp.Reset()
		for _, t := range testIDs {
			mp.Add(t)
		}
		for _, t := range testIDs {
			mp.Add(t)
		}
	}
}
