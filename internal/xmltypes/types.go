// Package xmltypes contains common types marshaled to xml
package xmltypes

import "encoding/xml"

// StringWithZero is like string, but marshals the empty string as "0"
type StringWithZero string

// Get gets the value as a string
func (s StringWithZero) Get() string {
	if s == "" {
		return "0"
	}
	return string(s)
}

// Set sets the value from a string
func (s *StringWithZero) Set(v string) {
	if v == "0" {
		*s = ""
	} else {
		*s = StringWithZero(v)
	}
}

func (s StringWithZero) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(s.Get(), start)
}

func (s *StringWithZero) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	s.Set(value)
	return nil
}

// BoolAsString is like bool, but marshals as either "TRUE" or "FALSE"
type BoolAsString bool

// Get gets the value as an integer
func (b BoolAsString) Get() string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

// Set sets the BoolAsString to contain a string value
func (b *BoolAsString) Set(v string) {
	*b = (v == "TRUE")
}

func (b BoolAsString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(b.Get(), start)
}

func (b *BoolAsString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	b.Set(value)
	return nil
}

// BoolAsInt is like bool, but marshals as 0 (false) or 1 (true)
type BoolAsInt bool

// Get gets the value as an integer
func (b BoolAsInt) Get() int {
	if b {
		return 1
	}
	return 0
}

// Set sets the BoolAsInt to contain an integer value
func (b *BoolAsInt) Set(v int) {
	*b = (v != 0)
}

func (b BoolAsInt) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(b.Get(), start)
}

func (b *BoolAsInt) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value int
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	b.Set(value)
	return nil
}
