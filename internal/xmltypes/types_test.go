// Package xmltypes contains common types marshaled to xml
package xmltypes

// cspell:words xmltypes

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func PrintMarshal(x any) {
	value, err := xml.Marshal(x)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(value))
}

func MustUnmarshal(dst any, src string) {
	err := xml.Unmarshal([]byte(src), dst)
	if err != nil {
		panic(err)
	}
}

func ExampleStringWithZero() {
	var s StringWithZero

	MustUnmarshal(&s, "<StringWithZero>0</StringWithZero>")
	fmt.Println(s)
	PrintMarshal(s)

	MustUnmarshal(&s, "<StringWithZero>hello world</StringWithZero>")
	fmt.Println(s)
	PrintMarshal(s)

	// Output: <StringWithZero>0</StringWithZero>
	// hello world
	// <StringWithZero>hello world</StringWithZero>
}

func ExampleStringWithZero_marshal() {
	{
		s := StringWithZero("")
		bytes, _ := xml.Marshal(s)
		fmt.Println(string(bytes))
	}

	{
		b := StringWithZero("hello world")
		bytes, _ := xml.Marshal(b)
		fmt.Println(string(bytes))
	}

	// Output: <StringWithZero>0</StringWithZero>
	// <StringWithZero>hello world</StringWithZero>
}

func TestStringWithZero(t *testing.T) {
	tests := []struct {
		source string
		want   StringWithZero
	}{
		{
			source: "<StringWithZero>0</StringWithZero>",
			want:   "",
		},
		{
			source: "<StringWithZero>hello world</StringWithZero>",
			want:   "hello world",
		},
		{
			source: "<StringWithZero></StringWithZero>",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			var s StringWithZero
			if err := xml.Unmarshal([]byte(tt.source), &s); err != nil {
				t.Errorf("StringWithZero.UnmarshalXML() error = %v, wantErr %v", err, nil)
			}

			if s != tt.want {
				t.Errorf("StringWithZero got = %v, want %v", s, tt.want)
			}
		})
	}
}

func ExampleBoolAsString() {
	var b BoolAsString

	MustUnmarshal(&b, "<BoolAsString>FALSE</BoolAsString>")
	fmt.Println(b)
	PrintMarshal(b)

	MustUnmarshal(&b, "<BoolAsString>TRUE</BoolAsString>")
	fmt.Println(b)
	PrintMarshal(b)

	// Output: false
	// <BoolAsString>FALSE</BoolAsString>
	// true
	// <BoolAsString>TRUE</BoolAsString>
}

func ExampleBoolAsString_marshal() {
	{
		b := BoolAsString(false)
		bytes, _ := xml.Marshal(b)
		fmt.Println(string(bytes))
	}

	{
		b := BoolAsString(true)
		bytes, _ := xml.Marshal(b)
		fmt.Println(string(bytes))
	}

	// Output: <BoolAsString>FALSE</BoolAsString>
	// <BoolAsString>TRUE</BoolAsString>
}

func TestBoolAsString(t *testing.T) {
	tests := []struct {
		source string
		want   BoolAsString
	}{
		{
			source: "<BoolAsString>TRUE</BoolAsString>",
			want:   true,
		},
		{
			source: "<BoolAsString>FALSE</BoolAsString>",
			want:   false,
		},
		{
			source: "<BoolAsString>STUFF</BoolAsString>",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			var b BoolAsString
			if err := xml.Unmarshal([]byte(tt.source), &b); err != nil {
				t.Errorf("BoolAsString.UnmarshalXML() error = %v, wantErr %v", err, nil)
			}

			if b != tt.want {
				t.Errorf("BoolAsString got = %v, want %v", b, tt.want)
			}
		})
	}
}

func ExampleBoolAsInt() {
	var b BoolAsInt

	MustUnmarshal(&b, "<BoolAsInt>0</BoolAsInt>")
	fmt.Println(b)
	PrintMarshal(b)

	MustUnmarshal(&b, "<BoolAsInt>1</BoolAsInt>")
	fmt.Println(b)
	PrintMarshal(b)
	// Output: false
	// <BoolAsInt>0</BoolAsInt>
	// true
	// <BoolAsInt>1</BoolAsInt>
}

func ExampleBoolAsInt_marshal() {
	{
		b := BoolAsInt(false)
		bytes, _ := xml.Marshal(b)
		fmt.Println(string(bytes))
	}

	{
		b := BoolAsInt(true)
		bytes, _ := xml.Marshal(b)
		fmt.Println(string(bytes))
	}

	// Output: <BoolAsInt>0</BoolAsInt>
	// <BoolAsInt>1</BoolAsInt>
}

func TestBoolAsInt(t *testing.T) {
	tests := []struct {
		source string
		want   BoolAsInt
	}{
		{
			source: "<BoolAsInt>0</BoolAsInt>",
			want:   false,
		},
		{
			source: "<BoolAsInt>1</BoolAsInt>",
			want:   true,
		},
		{
			source: "<BoolAsInt>-1</BoolAsInt>",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			var b BoolAsInt
			if err := xml.Unmarshal([]byte(tt.source), &b); err != nil {
				t.Errorf("BoolAsInt.UnmarshalXML() error = %v, wantErr %v", err, nil)
			}

			if b != tt.want {
				t.Errorf("BoolAsInt got = %v, want %v", b, tt.want)
			}
		})
	}
}
