package exporter

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/huandu/go-sqlbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// SQL implements an exporter for storing data inside an sql database.
// TODO(twiesing): For now this only supports string-like fields.
type SQL struct {
	DB *sql.DB
	l  sync.Mutex
}

func (sql *SQL) Begin(bundle *pathbuilder.Bundle, count int64) error {
	sql.l.Lock()
	defer sql.l.Unlock()

	return sql.CreateTable(bundle) // create a table for the given bundle
}

const (
	uriField    = "uri"
	parentField = "parent"
	fieldPrefix = "field_"
)

func (*SQL) Table(bundle *pathbuilder.Bundle) string {
	return bundle.Group.ID
}

func (*SQL) Column(field pathbuilder.Field) string {
	return fieldPrefix + field.ID
}

func (sql *SQL) CreateTable(bundle *pathbuilder.Bundle) error {
	// build all the child tables first!
	for _, child := range bundle.ChildBundles {
		if err := sql.CreateTable(child); err != nil {
			return err
		}
	}

	// drop the table if it already exists
	if _, err := sql.DB.Exec("DROP TABLE IF EXISTS " + sql.Table(bundle) + ";"); err != nil {
		return err
	}

	// create a table with fields for every field, and the child field
	table := sqlbuilder.CreateTable(sql.Table(bundle)).IfNotExists()
	table.Define(uriField, "TEXT", "NOT NULL")
	table.Define(parentField, "TEXT")
	for _, field := range bundle.ChildFields {
		table.Define(sql.Column(field))
	}

	// build the table after the child table
	query, args := table.Build()
	_, err := sql.DB.Exec(query, args...)
	return err
}

func (sql *SQL) Add(bundle *pathbuilder.Bundle, entity *wisski.Entity) error {
	sql.l.Lock()
	defer sql.l.Unlock()

	return sql.addWithParent(bundle, "", entity)
}

// addWithParent adds an entity with an optional parent
func (sql *SQL) addWithParent(bundle *pathbuilder.Bundle, parent wisski.URI, entity *wisski.Entity) error {
	// find all the fields and values to insert
	var (
		columns []string
		values  []any
	)

	columns = append(columns, uriField)
	values = append(values, string(entity.URI))

	if parent != "" {
		columns = append(columns, parentField)
		values = append(values, string(parent))
	}

	var builder strings.Builder
	for field, fvalues := range entity.Fields {
		if len(fvalues) == 0 {
			continue
		}

		for _, v := range fvalues {
			fmt.Fprintf(&builder, "%v,", v.Value)
		}

		columns = append(columns, sql.Column(bundle.Field(field)))
		values = append(values, builder.String()[:builder.Len()-1]) // trim trailing comma
		builder.Reset()
	}

	// build what to insert
	insert := sqlbuilder.InsertInto(sql.Table(bundle))
	insert.Cols(columns...)
	insert.Values(values...)

	// perform the insert
	query, args := insert.Build()
	_, err := sql.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	// insert all the children
	for name, children := range entity.Children {
		bundle := bundle.Bundle(name)
		for _, child := range children {
			if err := sql.addWithParent(bundle, entity.URI, &child); err != nil {
				return err
			}
		}
	}

	return nil
}

func (sql *SQL) End(bundle *pathbuilder.Bundle) error {
	return nil // no-op
}

func (sql *SQL) Close() error {
	return sql.DB.Close() // close the databas
}
