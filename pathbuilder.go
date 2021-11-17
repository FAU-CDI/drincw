package drincw

import (
	"encoding/xml"
)

type PathbuilderInterface struct {
	XMLName xml.Name `xml:"pathbuilderinterface"`
	Paths   []Path   `xml:"path"`
}

type Path struct {
	XMLName xml.Name `xml:"path"`

	ID      string    `xml:"id"`
	Weight  int       `xml:"weight"`
	Enabled BoolAsInt `xml:"enabled"`

	GroupID ZeroString `xml:"group_id"`
	Bundle  string     `xml:"bundle"`

	Field     string `xml:"field"`
	FieldType string `xml:"fieldtype"`

	DisplayWidget   string `xml:"displaywidget"`
	FormatterWidget string `xml:"formatterwidget"`

	Cardinality int `xml:"cardinality"`

	FieldTypeInformative string `xml:"field_type_informative"`

	PathArray string `xml:"path_array"`

	DatatypeProperty string `xml:"datatype_property"`

	ShortName string `xml:"short_name"`
	Disamb    int    `xml:"disam"`

	Description string `xml:"description"`
	UUID        string `xml:"uuid"`

	IsGroup BoolAsInt `xml:"is_group"`

	Name string `xml:"name"`
}
