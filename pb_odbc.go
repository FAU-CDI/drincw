package drincw

import "encoding/xml"

// ODBCServer represents a Server declaration for ODBC
type ODBCServer struct {
	XMLName xml.Name `xml:"server"`

	URL      string `xml:"url"`
	Database string `xml:"database"`
	Port     int    `xml:"port"`
	User     string `xml:"user"`
	Password string `xml:"password"`

	Tables []ODBCTable
}

// TableByID returns the table in this server with the provided main bundle id
// If no such table exists, returns an empty ODBCTable.
func (server ODBCServer) TableByID(mainBundleID string) ODBCTable {
	for _, table := range server.Tables {
		if table.MainBundleID() == mainBundleID {
			return table
		}
	}
	return ODBCTable{}
}

func (pb Pathbuilder) ODBC() (s ODBCServer) {
	s.URL = "localhost"
	s.Database = ""
	s.Port = 3306
	s.User = ""
	s.Password = ""

	bundles := pb.Bundles()
	s.Tables = make([]ODBCTable, len(bundles))
	for i, b := range bundles {
		s.Tables[i] = b.odbcTable()
	}
	return
}

type ODBCTable struct {
	XMLName xml.Name `xml:"table"`

	Select string `xml:"select"`
	Name   string `xml:"name"`

	Append    string           `xml:"append"`
	Delimiter string           `xml:"delimiter"`
	ID        string           `xml:"id"`
	Trim      xmlBoolTrueFalse `xml:"trim"`

	Row struct {
		ODBCBundlesAndFields
	} `xml:"row"`
}

func (bundle Bundle) odbcTable() (t ODBCTable) {
	t.Select = "*" // TODO: Generate something here
	t.Name = bundle.Group.ID

	t.Append = ""
	t.Delimiter = ";"
	t.ID = "id"
	t.Trim = true

	t.Row.ODBCBundlesAndFields.Bundles = []ODBCBundle{bundle.odbcBundle()}

	return
}

// MainBundleID returns the main bundle id corresponding to this table
func (table ODBCTable) MainBundleID() string {
	// if there are no bundles, return
	if len(table.Row.Bundles) == 0 {
		return ""
	}

	// id of the first bundle
	return table.Row.Bundles[0].ID
}

type ODBCBundle struct {
	XMLName xml.Name `xml:"bundle"`

	ID      string `xml:"id,attr"`
	Comment string `xml:",comment"`

	ODBCBundlesAndFields
}

func (bundle Bundle) odbcBundle() (b ODBCBundle) {
	b.ID = bundle.Group.Bundle
	b.Comment = " " + bundle.Group.Name + " "

	b.ODBCBundlesAndFields = bundle.odbcBundlesAndFields()

	return
}

type ODBCField struct {
	XMLName xml.Name `xml:"field"`

	ID string `xml:"id,attr"`

	Comment   string `xml:",comment"`
	FieldName string `xml:"fieldname"`
}

func (field Field) odbcField() (f ODBCField) {
	f.ID = field.Field
	f.FieldName = field.ID
	f.Comment = " " + field.Name + " "
	return
}

type ODBCBundlesAndFields struct {
	Fields  []ODBCField
	Bundles []ODBCBundle
}

func (bundle Bundle) odbcBundlesAndFields() (b ODBCBundlesAndFields) {
	fields := bundle.Fields()
	b.Fields = make([]ODBCField, len(fields))
	for i, f := range fields {
		b.Fields[i] = f.odbcField()
	}

	bundles := bundle.Bundles()
	b.Bundles = make([]ODBCBundle, len(bundles))
	for i, bb := range bundles {
		b.Bundles[i] = bb.odbcBundle()
	}

	return
}
