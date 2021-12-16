package sql

import (
	"errors"
	"fmt"
)

// Selector provides means of selecting a value from an sql table.
//
//
// Selectors are intentionally intransparent to the caller; they should only be accessed using the
// MarshalSelector and UnmarshalSelector methods.
//
// Internally any selector type takes a pointer receiver.
type Selector interface {
	// name must return the name of this selector
	//
	// this method must take a pointer receiver.
	name() Identifier

	// fields returns the fields used for unmarshaling and marshaling this selector.
	//
	// A field must either be of the following forms:
	// - "$StructField" where StructField is a field of the underlying struct literal.
	// - any string value not starting with '$', which will be assumed to occur literally in the source.
	//
	// this method must take a pointer receiver.
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

var errSelectorInvalidIdentifier = errors.New("Selector: invalid identifier")
var errSelectorNoAppend = errors.New("Selector: no append")

// ColumnSelector is a selector that selects a single column from the main sql table
type ColumnSelector struct {
	Column Identifier
}

func (*ColumnSelector) name() Identifier {
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

// JoinSelector selects Column from a secondary Table using a (left) join on equality of OurKey and TheirKey
type JoinSelector struct {
	Column Identifier

	Table Identifier

	OurKey   Identifier
	TheirKey Identifier
}

func (*JoinSelector) name() Identifier {
	return "join"
}

func (*JoinSelector) fields() []string {
	return []string{"$Column", "from", "$Table", "on", "$OurKey", "$TheirKey"}
}

func (j JoinSelector) selectExpression(table Identifier, temp IdentifierFactory) (string, error) {
	return fmt.Sprintf("%q.%q", temp.Get(""), j.Column), nil
}

func (j JoinSelector) appendStatement(table Identifier, temp IdentifierFactory) (string, error) {
	theirTable := Identifier(j.Table)
	theirKey := Identifier(j.TheirKey)
	tempTable := temp.Get("")
	ourTable := Identifier(table)
	ourKey := Identifier(j.OurKey)
	return fmt.Sprintf("LEFT JOIN %q AS %q ON %q.%q = %q.%q", theirTable, tempTable, ourTable, ourKey, tempTable, theirKey), nil
}

// Many2ManySelector selects a many2many relation.
type Many2ManySelector struct {
	Column Identifier
	Table  Identifier

	Through Identifier

	TheirKey        Identifier
	TheirThroughKey Identifier
	OurThroughKey   Identifier
	OurKey          Identifier
}

func (*Many2ManySelector) name() Identifier {
	return "many2many"
}

func (*Many2ManySelector) fields() []string {
	return []string{"$Column", "from", "$Table", "through", "$Through", "on", "$TheirKey", "$TheirThroughKey", "$OurThroughKey", "$OurKey"}
}

func (m Many2ManySelector) selectExpression(table Identifier, temp IdentifierFactory) (string, error) {
	through := temp.Get("through")
	throughValue := temp.Get("through_value")

	return fmt.Sprintf("%q.%q", through, throughValue), nil
}

func (m Many2ManySelector) appendStatement(table Identifier, temp IdentifierFactory) (string, error) {
	through := temp.Get("through")
	throughID := temp.Get("through_id")
	throughValue := temp.Get("through_value")

	throughSubquery := fmt.Sprintf(
		"SELECT %q.%q AS %q, GROUP_CONCAT(%q.%q SEPARATOR \"%s\") AS %q FROM %q LEFT JOIN %q ON %q.%q = %q.%q GROUP BY %q.%q",
		m.Through, m.OurThroughKey, throughID,
		m.Table, m.Column, ";", throughValue,
		m.Through, m.Table,
		m.Through, m.TheirThroughKey,
		m.Table, m.TheirKey,
		m.Table, m.OurKey,
	)

	return fmt.Sprintf("LEFT JOIN (%s) AS %q ON %q.%q = %q.%q", throughSubquery, through, through, throughID, table, m.OurKey), nil
}
