package drincw

import "encoding/xml"

// ZeroString is like string, but marshals the empty string as "0" to xml.
type ZeroString string

// Get gets the value as a string
func (s ZeroString) Get() string {
	if s == "" {
		return "0"
	}
	return string(s)
}

// Set sets the value from a string
func (s *ZeroString) Set(v string) {
	if v == "0" {
		*s = ""
	} else {
		*s = ZeroString(v)
	}
}

func (s ZeroString) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(s.Get(), start)
}

func (s *ZeroString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	s.Set(value)
	return nil
}

// BoolAsInt is like bool, but represents values as a 0 or a 1 when serializing to xml.
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
