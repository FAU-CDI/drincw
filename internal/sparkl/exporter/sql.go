package exporter

import (
	"database/sql"
	"errors"
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

	BatchSize   int // BatchSize for top-level bundles
	MaxQueryVar int // Maximum number of query variables (overrides BatchSize)

	Separator string // Seperator for database multi-valued fields

	batches map[string][]wisski.Entity

	l sync.Mutex
}

func (sql *SQL) Begin(bundle *pathbuilder.Bundle, count int64) error {
	sql.l.Lock()
	defer sql.l.Unlock()

	if sql.batches == nil {
		sql.batches = make(map[string][]wisski.Entity)
	}

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

	sql.batches[bundle.Group.ID] = append(sql.batches[bundle.Group.ID], *entity)
	if len(sql.batches) > sql.BatchSize-1 {
		sql.flushBatches(bundle)
	}

	return nil
}

func (sql *SQL) flushBatches(bundle *pathbuilder.Bundle) error {
	err := sql.doInserts(bundle, "", sql.batches[bundle.Group.ID])
	sql.batches[bundle.Group.ID] = nil
	return err
}

var nullString sql.NullString

const maxVariables = 999

var errSQLInsufficientVars = errors.New("Insufficient query variables")

// flushBuffers flushes the buffers and performs an actual insert
func (sql *SQL) doInserts(bundle *pathbuilder.Bundle, parent wisski.URI, entities []wisski.Entity) error {

	// find all the columns to insert
	var columns []string
	columns = append(columns, uriField)
	if parent != "" {
		columns = append(columns, parentField)
	}
	for _, field := range bundle.Fields() {
		columns = append(columns, sql.Column(field))
	}

	// compute the maximal chunk size
	chunkSize := sql.MaxQueryVar / len(columns)
	if chunkSize == 0 {
		return errSQLInsufficientVars
	}

	// iterate over each chunk (for which there are sufficient variables)
	var builder strings.Builder
	for i := 0; i < len(entities); i += chunkSize {
		insert := sqlbuilder.InsertInto(sql.Table(bundle))
		insert.Cols(columns...)

		// compute the true chunk bounds
		var (
			chunkStart = i
			chunkEnd   = i + chunkSize
		)
		if chunkEnd > len(entities) {
			chunkEnd = len(entities)
		}

		for _, entity := range entities[chunkStart:chunkEnd] {
			values := make([]any, 1, len(columns))
			values[0] = entity.URI

			if parent != "" {
				values = append(values, string(parent))
			}

			for _, field := range bundle.Fields() {
				fvalues := entity.Fields[field.ID]
				if len(fvalues) == 0 {
					values = append(values, nullString)
					continue
				}
				for _, v := range fvalues {
					fmt.Fprintf(&builder, "%v,", v.Value)
				}
				values = append(values, builder.String()[:builder.Len()-1])
				builder.Reset()
			}

			insert.Values(values...)
		}

		// perform the actual insert!
		query, args := insert.Build()
		_, err := sql.DB.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	// perform the insert of all children
	for _, bundle := range bundle.ChildBundles {
		for _, entity := range entities {
			children := entity.Children[bundle.Group.ID]
			for i := 0; i < len(children); i += sql.BatchSize {
				// compute the true batch bounds
				var (
					batchStart = i
					batchEnd   = i + sql.BatchSize
				)
				if batchEnd > len(children) {
					batchEnd = len(children)
				}
				if err := sql.doInserts(bundle, entity.URI, children[batchStart:batchEnd]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (sql *SQL) End(bundle *pathbuilder.Bundle) error {
	sql.l.Lock()
	defer sql.l.Unlock()

	return sql.flushBatches(bundle)
}

func (sql *SQL) Close() error {
	return sql.DB.Close() // close the databas
}
