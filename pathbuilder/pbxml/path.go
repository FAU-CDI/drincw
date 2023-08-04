package pbxml

// cspell:words pathbuilder

import (
	"encoding/xml"

	"github.com/FAU-CDI/drincw/pathbuilder"
	"github.com/FAU-CDI/drincw/pkg/xmltypes"
)

// path represents the "path" element of pathbuilder xml
type path struct {
	XMLName xml.Name `xml:"path"`

	ID      string             `xml:"id"`
	Weight  int                `xml:"weight"`
	Enabled xmltypes.BoolAsInt `xml:"enabled"`

	GroupID xmltypes.StringWithZero `xml:"group_id"`
	Bundle  string                  `xml:"bundle"`

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

	IsGroup xmltypes.BoolAsInt `xml:"is_group"`

	Name string `xml:"name"`
}

func newPath(path pathbuilder.Path) (x path) {
	x.ID = path.ID
	x.Weight = path.Weight
	x.Enabled = xmltypes.BoolAsInt(path.Enabled)
	x.GroupID = xmltypes.StringWithZero(path.GroupID)
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
	x.IsGroup = xmltypes.BoolAsInt(path.IsGroup)
	x.Name = path.Name

	return
}

func (x path) Path() (p pathbuilder.Path) {
	p.ID = x.ID
	p.Weight = x.Weight
	p.Enabled = bool(x.Enabled)
	p.GroupID = string(x.GroupID)
	p.Bundle = x.Bundle
	p.Field = x.Field
	p.FieldType = x.FieldType
	p.DisplayWidget = x.DisplayWidget
	p.FormatterWidget = x.FormatterWidget
	p.Cardinality = x.Cardinality
	p.FieldTypeInformative = x.FieldTypeInformative
	p.PathArray = x.PathArray
	p.DatatypeProperty = x.DatatypeProperty
	p.ShortName = x.ShortName
	p.Disamb = x.Disamb
	p.Description = x.Description
	p.UUID = x.UUID
	p.IsGroup = bool(x.IsGroup)
	p.Name = x.Name
	return
}
