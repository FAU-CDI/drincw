// Package xmltypes contains several types that encode differently in XML.
//
// Each type T in this package implements the Type interface.
package xmltypes

// cspell:words xmltypes

import "encoding/xml"

// Type represents a type represented as X inside of xml.
type Type[X any] interface {
	xml.Marshaler
	xml.Unmarshaler

	// Get the underlying type of this value
	Get() X

	// Set sets the value of this type
	Set(v X)
}

// check that types in this package actually implement Type
var (
	_ Type[string] = (*StringWithZero)(nil)
	_ Type[string] = (*BoolAsString)(nil)
	_ Type[int]    = (*BoolAsInt)(nil)
)

// StringWithZero is like string, but marshals the empty string as "0".
type StringWithZero string

func (s StringWithZero) Get() string {
	if s == "" {
		return "0"
	}
	return string(s)
}

func (s *StringWithZero) Set(v string) {
	if v == "0" {
		*s = ""
		return
	}

	*s = StringWithZero(v)
}

func (s StringWithZero) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[string](&s, e, start)
}

func (s *StringWithZero) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[string](s, d, start)
}

// BoolAsInt is a boolean that is marshaled as a string in xml.
// "TRUE" represents true, any other string represents false.
type BoolAsString bool

func (b BoolAsString) Get() string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func (b *BoolAsString) Set(v string) {
	*b = (v == "TRUE")
}

func (b BoolAsString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[string](&b, e, start)
}

func (b *BoolAsString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[string](b, d, start)
}

// BoolAsInt is a boolean that is marshaled as an integer in xml.
// 0 represents false, any other number represents true.
type BoolAsInt bool

// Get returns this boolean as an integer
func (b BoolAsInt) Get() int {
	if b {
		return 1
	}
	return 0
}

func (b *BoolAsInt) Set(v int) {
	*b = (v != 0)
}

func (b BoolAsInt) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[int](&b, e, start)
}

func (b *BoolAsInt) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[int](b, d, start)
}

// marshal and unmarshal implement xml.Marshal and xml.Unmarshal respectively.

func marshal[X any](w Type[X], e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(w.Get(), start)
}

func unmarshal[X any](w Type[X], d *xml.Decoder, start xml.StartElement) error {
	var value X
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	w.Set(value)
	return nil
}
