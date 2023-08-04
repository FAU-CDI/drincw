// Package odbc provides facilities for odbc declarations
package odbc

// cspell:words odbc pathbuilder

import (
	"encoding/xml"

	"github.com/FAU-CDI/drincw/pathbuilder"
)

// Server represents an odbc server implementation.
// It is the main interface to ODBC.
//
// It can be passed to xml.Marshal and xml.Unmarshal.
type Server struct {
	XMLName xml.Name `xml:"server"`

	URL      string `xml:"url"`
	Database string `xml:"database"`
	Port     int    `xml:"port"`
	User     string `xml:"user"`
	Password string `xml:"password"`

	Tables []Table `xml:"table"`
}

// NewServer generates a new server from a pathbuilder
func NewServer(pb pathbuilder.Pathbuilder) (s Server) {
	s.URL = "localhost"
	s.Database = ""
	s.Port = 3306
	s.User = ""
	s.Password = ""

	bundles := pb.Bundles()
	s.Tables = make([]Table, len(bundles))
	for i, b := range bundles {
		s.Tables[i] = newTable(*b)
	}
	return
}

// TableByID returns the table in this server with the provided main bundle id
// If no such table exists, returns an empty ODBCTable.
func (server Server) TableByID(mainBundleID string) Table {
	for _, table := range server.Tables {
		if table.MainBundleID() == mainBundleID {
			return table
		}
	}
	return Table{}
}
