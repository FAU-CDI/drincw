package imap

import "fmt"

func ExampleID() {

	// create a new id -- which isn't valid
	var id ID
	fmt.Println(id)
	fmt.Println(id.Valid())

	// increment the id -- it is now valid
	fmt.Println(id.Inc())
	fmt.Println(id.Valid())

	// compare it to some other id -- it is no longer valid
	other := ID([1]uint64{10})
	fmt.Println(id.Less(other))

	// Output: 0
	// false
	// 1
	// true
	// true
}
