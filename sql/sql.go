// Package sql provides facilities for mapping a pathbuilder and odbc importer to sql statements.
package sql

import (
	"fmt"

	"github.com/tkw1536/FAU-CDI/drincw"
)

// ForTable generates an sql statement used by the importer with the given table
func ForTable(table drincw.ODBCTable) string {
	id := Identifier(table.ID)
	name := Identifier(table.Name)

	var sSelect string
	if table.Select != "" {
		sSelect = ", " + table.Select
	}

	var append string
	if table.Append != "" {
		append = " " + table.Append
	}

	return fmt.Sprintf("SELECT %q.%q as %q%s FROM %q%s", name, id, Identifier("id"), sSelect, name, append)
}
