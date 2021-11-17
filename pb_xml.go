package drincw

import (
	"encoding/xml"
	"errors"
	"strings"
)

// XMLPathbuilder represents XML corresponding to a pathbuilder
type XMLPathbuilder struct {
	XMLName xml.Name  `xml:"pathbuilderinterface"`
	Paths   []XMLPath `xml:"path"`
}

func (pb Pathbuilder) XML() (x XMLPathbuilder) {
	paths := pb.Paths()
	x.Paths = make([]XMLPath, len(paths))
	for i, p := range paths {
		x.Paths[i] = p.XML()
	}
	return
}

type XMLPath struct {
	XMLName xml.Name `xml:"path"`

	ID      string    `xml:"id"`
	Weight  int       `xml:"weight"`
	Enabled xmlBool01 `xml:"enabled"`

	GroupID xmlStringWith0 `xml:"group_id"`
	Bundle  string         `xml:"bundle"`

	Field     string `xml:"field"`
	FieldType string `xml:"fieldtype"`

	DisplayWidget   string `xml:"displaywidget"`
	FormatterWidget string `xml:"formatterwidget"`

	Cardinality int `xml:"cardinality"`

	FieldTypeInformative string `xml:"field_type_informative"`

	PathArray xmlPathArray `xml:"path_array"`

	DatatypeProperty string `xml:"datatype_property"`

	ShortName string `xml:"short_name"`
	Disamb    int    `xml:"disam"`

	Description string `xml:"description"`
	UUID        string `xml:"uuid"`

	IsGroup xmlBool01 `xml:"is_group"`

	Name string `xml:"name"`
}

func (path Path) XML() (x XMLPath) {
	x.ID = path.ID
	x.Weight = path.Weight
	x.Enabled = xmlBool01(path.Enabled)
	x.GroupID = xmlStringWith0(path.GroupID)
	x.Bundle = path.Bundle
	x.Field = path.Field
	x.FieldType = path.FieldType
	x.DisplayWidget = path.DisplayWidget
	x.FormatterWidget = path.FormatterWidget
	x.Cardinality = path.Cardinality
	x.FieldTypeInformative = path.FieldTypeInformative
	x.PathArray = path.PathArray
	x.DatatypeProperty = path.DatatypeProperty
	x.ShortName = path.ShortName
	x.Disamb = path.Disamb
	x.Description = path.Description
	x.UUID = path.UUID
	x.IsGroup = xmlBool01(path.IsGroup)
	x.Name = path.Name

	return
}

// xmlStringWith0 is like string, but marshals the empty string as "0" to xml.
type xmlStringWith0 string

// Get gets the value as a string
func (s xmlStringWith0) Get() string {
	if s == "" {
		return "0"
	}
	return string(s)
}

// Set sets the value from a string
func (s *xmlStringWith0) Set(v string) {
	if v == "0" {
		*s = ""
	} else {
		*s = xmlStringWith0(v)
	}
}

func (s xmlStringWith0) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(s.Get(), start)
}

func (s *xmlStringWith0) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	s.Set(value)
	return nil
}

// xmlBoolTrueFalse is like bool, but writes either "TRUE" or "FALSE"
type xmlBoolTrueFalse bool

// Get gets the value as an integer
func (b xmlBoolTrueFalse) Get() string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

// Set sets the BoolAsString to contain a string value
func (b *xmlBoolTrueFalse) Set(v string) {
	*b = (v == "TRUE")
}

func (b xmlBoolTrueFalse) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(b.Get(), start)
}

func (b *xmlBoolTrueFalse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	b.Set(value)
	return nil
}

// xmlBool01 is like bool, but represents values as a 0 or a 1 when serializing to xml.
type xmlBool01 bool

// Get gets the value as an integer
func (b xmlBool01) Get() int {
	if b {
		return 1
	}
	return 0
}

// Set sets the BoolAsInt to contain an integer value
func (b *xmlBool01) Set(v int) {
	*b = (v != 0)
}

func (b xmlBool01) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(b.Get(), start)
}

func (b *xmlBool01) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value int
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}
	b.Set(value)
	return nil
}

// xmlPathArray represents a set of paths
type xmlPathArray []string

func (paths xmlPathArray) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	var isX bool
	var startPath xml.StartElement
	for _, path := range paths {
		isX = !isX
		if isX {
			startPath.Name.Local = "x"
		} else {
			startPath.Name.Local = "y"
		}

		if err := encoder.EncodeToken(startPath); err != nil {
			return err
		}

		if err := encoder.EncodeToken(xml.CharData([]byte(path))); err != nil {
			return err
		}

		if err := encoder.EncodeToken(startPath.End()); err != nil {
			return err
		}
	}

	return encoder.EncodeToken(start.End())
}

var errPathArrayNoCharData = errors.New("PathArray: Missing CharData for path")
var errPathArrayNoOpen = errors.New("PathArray: Expected opening <x> or <y>")
var errPathArrayNoClose = errors.New("PathArray: Expected closing </x> or </y>")
var errPathArrayInvalid = errors.New("PathArray: Received invalid token")

func (paths *xmlPathArray) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	nextOpen := "x"
	var expectText bool
	var expectClose bool

	results := make([]string, 0)
readloop:
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		// ignore space only tokens
		if text, isText := token.(xml.CharData); isText && strings.TrimSpace(string(text)) == "" {
			continue
		}

		// ignore comments
		if _, isComment := token.(xml.Comment); isComment {
			continue
		}

		switch {
		case expectText:
			text, isText := token.(xml.CharData)
			if !isText {
				return errPathArrayNoCharData
			}

			results = append(results, string(text))

			expectText = false
			expectClose = true
		case expectClose:
			_, isClose := token.(xml.EndElement)
			if !isClose {
				return errPathArrayNoClose
			}
			expectText = false
			expectClose = false
		default:
			open, isOpen := token.(xml.StartElement)
			if isOpen {
				if open.Name.Local != nextOpen {
					return errPathArrayNoOpen
				}
				if nextOpen == "x" {
					nextOpen = "y"
				} else {
					nextOpen = "x"
				}
				expectText = true
				expectClose = false
				continue
			}

			_, isClose := token.(xml.EndElement)
			if isClose {
				break readloop
			}
			return errPathArrayInvalid
		}
	}

	*paths = xmlPathArray(results)
	return nil
}
