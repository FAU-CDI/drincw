// Package xmltypes contains common types marshaled to xml
package xmltypes

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func ExampleStringWithZero_Set() {
	var s StringWithZero

	s.Set("0")
	fmt.Println(s)
	fmt.Println(s.Get())

	s.Set("hello world")
	fmt.Println(s)
	fmt.Println(s.Get())

	// Output: 0
	// hello world
	// hello world
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

func ExampleBoolAsString_Set() {
	var b BoolAsString

	b.Set("FALSE")
	fmt.Println(b)
	fmt.Println(b.Get())

	b.Set("TRUE")
	fmt.Println(b)
	fmt.Println(b.Get())

	// Output: false
	// FALSE
	// true
	// TRUE
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

func ExampleBoolAsInt_Set() {
	var b BoolAsInt

	b.Set(0)
	fmt.Println(b)
	fmt.Println(b.Get())

	b.Set(1)
	fmt.Println(b)
	fmt.Println(b.Get())
	// Output: false
	// 0
	// true
	// 1
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
