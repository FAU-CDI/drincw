package imap

import (
	"fmt"
	"strconv"
	"testing"
)

func ExampleIMap() {

	var mp IMap[string]
	mp.Reset(MemoryEngine[string]{})

	lid := func(prefix string) func(id ID, err error) {
		return func(id ID, err error) {
			fmt.Println(prefix, id, err)
		}
	}

	lstr := func(prefix string) func(value string, err error) {
		return func(value string, err error) {
			fmt.Println(prefix, value, err)
		}
	}

	lid("add")(mp.Add("hello"))
	lid("add")(mp.Add("world"))
	lid("add")(mp.Add("earth"))

	lid("add<again>")(mp.Add("hello"))
	lid("add<again>")(mp.Add("world"))
	lid("add<again>")(mp.Add("earth"))

	lid("get")(mp.Forward("hello"))
	lid("get")(mp.Forward("world"))
	lid("get")(mp.Forward("earth"))

	lstr("reverse")(mp.Reverse([1]uint64{1}))
	lstr("reverse")(mp.Reverse([1]uint64{2}))
	lstr("reverse")(mp.Reverse([1]uint64{3}))

	mp.MarkIdentical("earth", "world")

	lstr("reverse<again>")(mp.Reverse([1]uint64{1}))
	lstr("reverse<again>")(mp.Reverse([1]uint64{3}))

	lid("add<again>")(mp.Add("hello"))
	lid("add<again>")(mp.Add("world"))
	lid("add<again>")(mp.Add("earth"))

	// Output: add 1 <nil>
	// add 2 <nil>
	// add 3 <nil>
	// add<again> 1 <nil>
	// add<again> 2 <nil>
	// add<again> 3 <nil>
	// get 1 <nil>
	// get 2 <nil>
	// get 3 <nil>
	// reverse hello <nil>
	// reverse world <nil>
	// reverse earth <nil>
	// reverse<again> hello <nil>
	// reverse<again> earth <nil>
	// add<again> 1 <nil>
	// add<again> 3 <nil>
	// add<again> 3 <nil>
}

var testIDs [10000]string

func init() {
	for i := 0; i < 10000; i++ {
		testIDs[i] = strconv.Itoa(i)
	}
}

func BenchmarkIMap(b *testing.B) {
	/*
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
	*/
}
