package odbc

import (
	"encoding/xml"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/xmltypes"
)

type Table struct {
	XMLName xml.Name `xml:"table"`

	Select string `xml:"select"`
	Name   string `xml:"name"`

	Append    string                `xml:"append"`
	Delimiter string                `xml:"delimiter"`
	ID        string                `xml:"id"`
	Trim      xmltypes.BoolAsString `xml:"trim"`

	Row struct {
		BundlesAndFields
	} `xml:"row"`
}

func newTable(bundle pathbuilder.Bundle) (t Table) {
	t.Select = "*" // TODO: Generate something here
	t.Name = bundle.Path.ID

	t.Append = ""
	t.Delimiter = ";"
	t.ID = "id"
	t.Trim = true

	t.Row.BundlesAndFields.Bundles = []Bundle{newBundle(bundle)}

	return
}

// MainBundleID returns the main bundle id corresponding to this table
func (table Table) MainBundleID() string {
	// if there are no bundles, return
	if len(table.Row.Bundles) == 0 {
		return ""
	}

	// id of the first bundle
	return table.Row.Bundles[0].ID
}

type Bundle struct {
	XMLName xml.Name `xml:"bundle"`

	ID      string `xml:"id,attr"`
	Comment string `xml:",comment"`

	BundlesAndFields
}

func newBundle(bundle pathbuilder.Bundle) (b Bundle) {
	b.ID = bundle.Path.Bundle
	b.Comment = " " + bundle.Path.Name + " "

	b.BundlesAndFields = newBundlesAndFields(bundle)
	return
}

type Field struct {
	XMLName xml.Name `xml:"field"`

	ID string `xml:"id,attr"`

	Comment   string `xml:",comment"`
	FieldName string `xml:"fieldname"`
}

func newField(field pathbuilder.Field) (f Field) {
	f.ID = field.Field
	f.FieldName = field.ID
	f.Comment = " " + field.Name + " "
	return
}

type BundlesAndFields struct {
	Fields  []Field
	Bundles []Bundle
}

func newBundlesAndFields(bundle pathbuilder.Bundle) (b BundlesAndFields) {
	fields := bundle.Fields()
	b.Fields = make([]Field, len(fields))
	for i, f := range fields {
		b.Fields[i] = newField(f)
	}

	bundles := bundle.Bundles()
	b.Bundles = make([]Bundle, len(bundles))
	for i, bb := range bundles {
		b.Bundles[i] = newBundle(*bb)
	}

	return
}
