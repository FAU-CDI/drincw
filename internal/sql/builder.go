package sql

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw/odbc"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"golang.org/x/exp/slices"
)

// Builder provides a correspondance between bundle ids and TableBuilder.
//
// New values should be created using make().
// The zero value does not cause panic(), but can not hold and correspondences
type Builder map[string]TableBuilder

// NewBuilder creates a new builder from a pathbuilder.
//
// Each bundle in the pathbuilder will correspond to a new TableBuilder.
// See BundleBuilder for details.
func NewBuilder(pb pathbuilder.Pathbuilder) Builder {
	bundles := pb.Bundles()
	b := make(map[string]TableBuilder, len(bundles))
	for _, bundle := range bundles {
		b[bundle.MachineName()] = NewTableBuilder(*bundle)
	}
	return b
}

// Apply updates the provided ODBC instance tables with correspondences provided within this Builder.
// Tables that do not have any correspondance will be removed from server.
func (b Builder) Apply(server *odbc.Server) error {
	tables := make([]odbc.Table, 0, len(server.Tables))
	orders := make(map[string]int, len(server.Tables))
	for _, table := range server.Tables {
		bb, ok := b[table.Name]
		if !ok {
			continue
		}
		orders[bb.TableName] = bb.Order
		if err := bb.Apply(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}

	// re-sort the tables by the provided order
	slices.SortStableFunc(tables, func(x, y odbc.Table) bool {
		return orders[x.Name] < orders[y.Name]
	})

	server.Tables = tables

	return nil
}

// TableBuilder provides facitilies to create sql statements for ODBC tables.
type TableBuilder struct {
	TableName string // name of the table to use
	ID        string // name of the column for ID
	Disinct   bool   // should we select distinct fields?
	Order     int    // order of different tables

	Fields map[string]Selector // Selectors for each bundle
}

// NewTableBuilder creates a new default TableBuilder for the given bundle.
//
// Each enabled field in the bundle will have a corresponding selector.
// Any further details are an implementation detail, and should not be relied upon by the caller.
func NewTableBuilder(bundle pathbuilder.Bundle) TableBuilder {
	tb := TableBuilder{}
	tb.TableName = bundle.MachineName()
	tb.ID = "id"

	fields := bundle.AllFields()
	tb.Fields = make(map[string]Selector, len(fields))
	for _, field := range fields {
		if !field.Enabled {
			continue
		}
		tb.Fields[field.MachineName()] = &ColumnSelector{Identifier(field.MachineName())}
	}

	return tb
}

// Apply updates the provided ODBC table with correspondences provided within this Builder.
//
// Bundles inside a table that do not have a corresponding sql in this TableBuilder will be removed.
func (tb TableBuilder) Apply(table *odbc.Table) error {
	table.Name = tb.TableName

	selectors := make(map[string]Selector)
	names := make(map[string]string)

	bundles := make([]odbc.Bundle, 0, len(table.Row.Bundles))
	for _, bundle := range table.Row.Bundles {
		if !tb.applyBundle(&bundle, selectors, names) {
			continue
		}
		bundles = append(bundles, bundle)
	}
	table.Row.Bundles = bundles

	fields := make([]odbc.Field, 0, len(table.Row.Fields))
	for _, field := range table.Row.Fields {
		if !tb.applyField(&field, selectors, names) {
			continue
		}
		fields = append(fields, field)
	}
	table.Row.Fields = fields

	var err error
	table.Select, table.Append, err = tb.build(selectors, names)
	if err != nil {
		return err
	}

	return nil
}

func (tb TableBuilder) applyBundle(bundle *odbc.Bundle, selectors map[string]Selector, names map[string]string) (ok bool) {
	fields := make([]odbc.Field, 0, len(bundle.Fields))
	for _, field := range bundle.Fields {
		if !tb.applyField(&field, selectors, names) {
			continue
		}
		fields = append(fields, field)
		ok = true
	}
	bundle.Fields = fields

	bundles := make([]odbc.Bundle, 0, len(bundle.Bundles))
	for _, bundle := range bundle.Bundles {
		if !tb.applyBundle(&bundle, selectors, names) {
			continue
		}
		bundles = append(bundles, bundle)
		ok = true
	}
	bundle.Bundles = bundles

	return
}

func (tb TableBuilder) applyField(field *odbc.Field, selectors map[string]Selector, names map[string]string) (ok bool) {
	selector, ok := tb.Fields[field.FieldName]
	if !ok { // field doesn't exist
		return false
	}

	// set a fieldname, fallback to ID
	if field.FieldName == "" {
		field.FieldName = field.ID
	}

	// store names and selectors of fields
	names[field.ID] = field.FieldName
	selectors[field.ID] = selector

	return true
}

// Build builds two sql strings for usage within the odbc importer for this table.
//
// The select statement contains a list of fields to be selected.
// The appen statement represents an abitrary sql statement that should be appened to the sql statement as a whole.
//
// Either SQL statement is escaped and can be safely inserted inside an sql statement.
func (tb TableBuilder) Build() (selectS, appendS string, err error) {
	return tb.build(tb.Fields, nil)
}

func (tb TableBuilder) build(fields map[string]Selector, names map[string]string) (selects, appends string, err error) {
	var selectorS, appendS []string

	// generate a consistent ordering for the fields
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	bbTable := Identifier(tb.TableName)

	// iterate over them
	for _, key := range keys {
		temp := IdentifierFactory(fmt.Sprintf("column_%s", key))

		// get the name of the column for the sql
		name := names[key]
		if name == "" {
			name = key
		}
		name = Identifier(names[key]).Quoted()

		s, err := fields[key].selectExpression(bbTable, temp)
		if err != nil {
			return "", "", err
		}
		selectorS = append(selectorS, fmt.Sprintf("%s as %s", s, name))

		a, err := fields[key].appendStatement(bbTable, temp)
		if err == errSelectorNoAppend {
			continue
		}
		if err != nil {
			return "", "", err
		}
		appendS = append(appendS, a)
	}

	selectPrefix := ""
	if tb.Disinct {
		selectPrefix = "DISTINCT "
	}

	return selectPrefix + strings.Join(selectorS, ", "), strings.Join(appendS, " "), nil
}
