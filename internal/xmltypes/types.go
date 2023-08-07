// Package xmltypes contains several types that encode differently in XML.
package xmltypes

// cspell:words xmltypes

import "encoding/xml"

// typ represents a type represented as X inside of xml
type typ[X any] interface {
	get[X]
	set[X]
}

// get represents a type that can be written to XML.
// It is represented as type X in XML.
type get[X any] interface {
	xml.Marshaler
	get() X
}

// set represents a type that can be read from XML.
// It is represented as type X in XML.
type set[X any] interface {
	xml.Unmarshaler
	set(v X)
}

// marshal and unmarshal implement xml.Marshal and xml.Unmarshal respectively.

func marshal[X any](w get[X], e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(w.get(), start)
}

func unmarshal[X any](w set[X], d *xml.Decoder, start xml.StartElement) error {
	var value X
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	w.set(value)
	return nil
}

// check that types in this package actually implement Type
var (
	_ typ[string] = (*StringWithZero)(nil)
	_ typ[string] = (*BoolAsString)(nil)
	_ typ[int]    = (*BoolAsInt)(nil)
)

// StringWithZero is like string, but marshals the empty string as "0".
type StringWithZero string

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (s StringWithZero) get() string {
	if s == "" {
		return "0"
	}
	return string(s)
}

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (s *StringWithZero) set(v string) {
	if v == "0" {
		*s = ""
		return
	}

	*s = StringWithZero(v)
}

func (s StringWithZero) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[string](s, e, start)
}

func (s *StringWithZero) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[string](s, d, start)
}

// BoolAsInt is a boolean that is marshaled as a string in xml.
// "TRUE" represents true, any other string represents false.
type BoolAsString bool

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (b BoolAsString) get() string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (b *BoolAsString) set(v string) {
	*b = (v == "TRUE")
}

func (b BoolAsString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[string](b, e, start)
}

func (b *BoolAsString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[string](b, d, start)
}

// BoolAsInt is a boolean that is marshaled as an integer in xml.
// 0 represents false, any other number represents true.
type BoolAsInt bool

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (b BoolAsInt) get() int {
	if b {
		return 1
	}
	return 0
}

//lint:ignore U1000 false positive (https://github.com/dominikh/go-tools/issues/1294)
func (b *BoolAsInt) set(v int) {
	*b = (v != 0)
}

func (b BoolAsInt) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return marshal[int](b, e, start)
}

func (b *BoolAsInt) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return unmarshal[int](b, d, start)
}
