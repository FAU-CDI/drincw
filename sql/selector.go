package sql

import (
	"errors"
	"fmt"
)

// Selector provides means of selecting a value from an sql table
type Selector interface {
	withName

	// fields returns the fields used for unmarshaling and marshaling this selector.
	//
	// A field must either be of the following forms:
	// - "$StructField" where StructField is a field of the underlying struct literal.
	// - any string value not starting with '$', which will be assumed to occur literally in the source.
	fields() []string

	// selectExpression generates an expression to insert into an sql select statement.
	// It will be used roughly like:
	//
	//   "SELECT " + selectExpression(table, temp) + " AS my_column FROM " + table
	//
	// table is the name of the primary table.
	// temp is the name of a temporary identifier that is guaranteed to be unique between different selectors.
	selectExpression(table Identifier, temp IdentifierFactory) (string, error)

	// appendStatment generates a statement that will be inserted at the end of the sql statement.
	// when err is
	// It will be used roughly like:
	//
	// "SELECT ... FROM ... " + appendStatement(table, temp)
	//
	// table is the name of the primary table.
	// temp is the name of a temporary identifier that is guaranteed to be unique between different selectors.
	appendStatement(table Identifier, temp IdentifierFactory) (string, error)
}

type withName interface {
	// name must return the name of this selector
	// it may not be implemented in a pointer type
	name() Identifier
}

var errSelectorInvalidIdentifier = errors.New("Selector: invalid identifier")
var errSelectorNoAppend = errors.New("Selector: no append")

// ColumnSelector selects a single Column from the main table
type ColumnSelector struct {
	Column Identifier
}

func (ColumnSelector) name() Identifier {
	return "column"
}

func (*ColumnSelector) fields() []string {
	return []string{"$Column"}
}

func (c ColumnSelector) selectExpression(table Identifier, temp IdentifierFactory) (string, error) {
	return fmt.Sprintf("%q.%q", table, c.Column), nil
}

func (c ColumnSelector) appendStatement(table Identifier, temp IdentifierFactory) (string, error) {
	return "", errSelectorNoAppend
}

// JoinSelector selects Column from Table using a (left) join on OurKey, TheirKey
type JoinSelector struct {
	Column Identifier

	Table Identifier

	OurKey   Identifier
	TheirKey Identifier
}

func (JoinSelector) name() Identifier {
	return "join"
}

func (*JoinSelector) fields() []string {
	return []string{"$Column", "from", "$Table", "on", "$OurKey", "$TheirKey"}
}

func (j JoinSelector) selectExpression(table Identifier, temp IdentifierFactory) (string, error) {
	return fmt.Sprintf("%q.%q", temp, j.Column), nil
}

func (j JoinSelector) appendStatement(table Identifier, temp IdentifierFactory) (string, error) {
	theirTable, ok := Identifier(j.Table).Escape()
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	theirKey, ok := Identifier(j.TheirKey).Escape()
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	tempTable, ok := Identifier(temp).Escape()
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	ourTable, ok := Identifier(table).Escape()
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	ourKey, ok := Identifier(j.OurKey).Escape()
	if !ok {
		return "", errSelectorInvalidIdentifier
	}

	return fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s", theirTable, tempTable, ourTable, ourKey, tempTable, theirKey), nil
}
