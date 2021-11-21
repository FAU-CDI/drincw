package sql

import (
	"errors"
	"fmt"
)

// Selector provides means of selecting a value from an sql table
type Selector interface {
	// unmarshalFields unmarshals this selector from a tokenized list of fields.
	unmarshalFields(fields []string) error

	// marshalFields marshals this selector into a tokenized list of fields.
	marshalFields() []string

	// selectExpression generates an expression to insert into an sql select statement.
	// It will be used roughly like:
	//
	//   "SELECT " + selectExpression(table, temp) + " AS my_column FROM " + table
	//
	// table is the name of the primary table.
	// temp is the name of a temporary identifier that is guaranteed to be unique between different selectors.
	selectExpression(table string, temp string) (string, error)

	// appendStatment generates a statement that will be inserted at the end of the sql statement.
	// when err is
	// It will be used roughly like:
	//
	// "SELECT ... FROM ... " + appendStatement(table, temp)
	//
	// table is the name of the primary table.
	// temp is the name of a temporary identifier that is guaranteed to be unique between different selectors.
	appendStatement(table string, temp string) (string, error)
}

var errSelectorInvalidIdentifier = errors.New("Selector: invalid identifier")
var errSelectorNoAppend = errors.New("Selector: no append")

// ColumnSelector selects a single Column from the main table
type ColumnSelector struct {
	Column string
}

func (c *ColumnSelector) unmarshalFields(fields []string) error {
	if len(fields) != 1 {
		return errors.New("column: exactly one argument required")
	}
	c.Column = fields[0]
	return nil
}

func (c ColumnSelector) marshalFields() []string {
	// "column <name>"
	return []string{c.Column}
}

func (c ColumnSelector) selectExpression(table string, temp string) (string, error) {
	table, ok := EscapeIdentifier(table)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	column, ok := EscapeIdentifier(c.Column)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	return fmt.Sprintf("%s.%s", table, column), nil
}

func (c ColumnSelector) appendStatement(table string, temp string) (string, error) {
	return "", errSelectorNoAppend
}

// JoinSelector selects Column from Table using a (left) join on OurKey, TheirKey
type JoinSelector struct {
	Column string

	Table string

	OurKey   string
	TheirKey string
}

func (j *JoinSelector) unmarshalFields(fields []string) error {
	if len(fields) != 6 {
		return errors.New("join: exactly six fields required")
	}
	j.Column = fields[0]
	if fields[1] != "from" {
		return errors.New("join: second field must be 'from'")
	}
	j.Table = fields[2]
	if fields[3] != "on" {
		return errors.New("join: fourth field must be 'on'")
	}
	j.OurKey = fields[4]
	j.TheirKey = fields[5]

	return nil
}

func (j JoinSelector) marshalFields() []string {
	return []string{j.Column, "from", j.Table, "on", j.OurKey, j.TheirKey}
}

func (j JoinSelector) selectExpression(table string, temp string) (string, error) {
	temp, ok := EscapeIdentifier(temp)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	column, ok := EscapeIdentifier(j.Column)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	return fmt.Sprintf("%s.%s", temp, column), nil
}

func (j JoinSelector) appendStatement(table string, temp string) (string, error) {
	theirTable, ok := EscapeIdentifier(j.Table)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	theirKey, ok := EscapeIdentifier(j.TheirKey)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	tempTable, ok := EscapeIdentifier(temp)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	ourTable, ok := EscapeIdentifier(table)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	ourKey, ok := EscapeIdentifier(j.OurKey)
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	return fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s", theirTable, tempTable, ourTable, ourKey, tempTable, theirKey), nil
}
