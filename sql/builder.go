package sql

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw"
)

// Builder maps bundle ids to BundleBuilders
type Builder map[string]BundleBuilder

func NewBuilder(pb drincw.Pathbuilder) Builder {
	bundles := pb.Bundles()
	builder := make(map[string]BundleBuilder, len(bundles))
	for _, bundle := range bundles {
		builder[bundle.Group.ID] = NewBundleBuilder(bundle)
	}
	return builder
}

func (builder Builder) Apply(server *drincw.ODBCServer) error {
	tables := make([]drincw.ODBCTable, 0, len(server.Tables))
	for _, table := range server.Tables {
		bb, ok := builder[table.Name]
		if !ok {
			continue
		}
		if err := bb.Apply(&table); err != nil {
			return err
		}
		tables = append(tables, table)
	}
	server.Tables = tables
	return nil
}

// BundleBuilder builds a mapping from an sql table to a set of fields
type BundleBuilder struct {
	TableName string // name of the table to use
	ID        string // name of the column for ID
	Disinct   bool   // should we select distinct fields?

	Fields map[string]Selector // Selectors for each bundle
}

func NewBundleBuilder(bundle *drincw.Bundle) BundleBuilder {
	builder := BundleBuilder{}
	builder.TableName = bundle.Group.ID
	builder.ID = "id"

	fields := bundle.AllFields()
	builder.Fields = make(map[string]Selector, len(fields))
	for _, field := range fields {
		builder.Fields[field.Path.ID] = &ColumnSelector{Identifier(field.Path.ID)}
	}

	return builder
}

type bundleBuilderJSON struct {
	TableName string            `json:"table"`
	ID        string            `json:"id"`
	Distinct  bool              `json:"distinct"`
	Fields    map[string]string `json:"fields"`
}

func (bb BundleBuilder) MarshalJSON() ([]byte, error) {
	jb := bundleBuilderJSON{
		TableName: bb.TableName,
		ID:        bb.ID,
		Distinct:  bb.Disinct,
		Fields:    make(map[string]string, len(bb.Fields)),
	}

	var err error
	for field, selector := range bb.Fields {
		jb.Fields[field], err = MarshalSelector(selector)
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(jb)
}

func (bb *BundleBuilder) UnmarshalJSON(data []byte) error {
	jb := bundleBuilderJSON{}
	if err := json.Unmarshal(data, &jb); err != nil {
		return err
	}

	bb.TableName = jb.TableName
	bb.ID = jb.ID
	bb.Disinct = jb.Distinct
	bb.Fields = make(map[string]Selector, len(jb.Fields))

	var err error
	for field, selector := range jb.Fields {
		bb.Fields[field], err = UnmarshalSelector(selector)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bb BundleBuilder) Apply(table *drincw.ODBCTable) error {
	table.Name = bb.TableName

	selectors := make(map[string]Selector)
	names := make(map[string]string)

	bundles := make([]drincw.ODBCBundle, 0, len(table.Row.Bundles))
	for _, bundle := range table.Row.Bundles {
		if !bb.ApplyBundle(&bundle, selectors, names) {
			continue
		}
		bundles = append(bundles, bundle)
	}
	table.Row.Bundles = bundles

	fields := make([]drincw.ODBCField, 0, len(table.Row.Fields))
	for _, field := range table.Row.Fields {
		if !bb.ApplyField(&field, selectors, names) {
			continue
		}
		fields = append(fields, field)
	}
	table.Row.Fields = fields

	var err error
	table.Select, table.Append, err = bb.build(selectors, names)
	if err != nil {
		return err
	}

	return nil
}

func (bb BundleBuilder) ApplyBundle(bundle *drincw.ODBCBundle, selectors map[string]Selector, names map[string]string) (ok bool) {
	fields := make([]drincw.ODBCField, 0, len(bundle.Fields))
	for _, field := range bundle.Fields {
		if !bb.ApplyField(&field, selectors, names) {
			continue
		}
		fields = append(fields, field)
		ok = true
	}
	bundle.Fields = fields

	bundles := make([]drincw.ODBCBundle, 0, len(bundle.Bundles))
	for _, bundle := range bundle.Bundles {
		if !bb.ApplyBundle(&bundle, selectors, names) {
			continue
		}
		bundles = append(bundles, bundle)
		ok = true
	}
	bundle.Bundles = bundles

	return
}

func (bb BundleBuilder) ApplyField(field *drincw.ODBCField, selectors map[string]Selector, names map[string]string) (ok bool) {
	selector, ok := bb.Fields[field.FieldName]
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

func (bb BundleBuilder) Build() (selects, appends string, err error) {
	return bb.build(bb.Fields, nil)
}

func (bb BundleBuilder) build(fields map[string]Selector, names map[string]string) (selects, appends string, err error) {
	var selectorS, appendS []string

	// generate a consistent ordering for the fields
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	bbTable := Identifier(bb.TableName)

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
	if bb.Disinct {
		selectPrefix = "DISTINCT "
	}

	return selectPrefix + strings.Join(selectorS, ", "), strings.Join(appendS, " "), nil
}

// ForTable returns the SQL Statement corresponding to a specific table
func ForTable(table drincw.ODBCTable) string {
	id := Identifier(table.ID)
	sSelect := ""
	if table.Select != "" {
		sSelect = ", " + table.Select
	}

	name := Identifier(table.Name)

	suffix := ""
	if table.Append != "" {
		suffix = " " + table.Append
	}
	return fmt.Sprintf("SELECT %q.%q as %q%s FROM %q%s", name, id, Identifier("id"), sSelect, name, suffix)
}
