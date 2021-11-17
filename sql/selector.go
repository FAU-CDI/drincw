package sql

import (
	"fmt"
	"strings"
)

// Selector represents an sql selector for a path
type Selector interface {
	// part to put into the select clause
	selectClause(table, name string) string
	// part to be put into the append clause
	appendClause(table, name string) string
}

const allowedChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghiklmnopqrstuvwxyz_"

// Escape escapes value as a column name
// TODO: This calls panic if unsafe, make it actually escape!
func Escape(value string) string {
	for _, r := range value {
		if !strings.ContainsRune(allowedChars, r) {
			panic("Unsafe rune " + string(r) + " in name " + value)
		}
	}
	return "`" + value + "`"
}

// NewSelect parses a new selector from line
func NewSelector(line string) Selector {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}

	switch fields[0] {
	case "raw":
		if len(fields) != 2 {
			return nil
		}
		return Raw(fields[1])
	case "column":
		if len(fields) != 2 {
			return nil
		}
		return Column(Escape(fields[1]))
	case "left-join":
		if len(fields) != 6 {
			return nil
		}
		return LeftJoin{
			Table: Escape(fields[1]),
			Alias: Escape(fields[2]),

			OurKey:     Escape(fields[3]),
			ForeignKey: Escape(fields[4]),

			ForeignColumn: Escape(fields[5]),
		}
	}
	return nil
}

type Column string

func (c Column) selectClause(table, name string) string {
	return fmt.Sprintf("%s.%s as %s", table, string(c), name)
}

func (c Column) appendClause(table, name string) string {
	return ""
}

type Raw string

func (ss Raw) selectClause(table, name string) string {
	return fmt.Sprintf("%s as %s", string(ss), name)
}

func (ss Raw) appendClause(table, name string) string {
	return ""
}

type LeftJoin struct {
	Table string // foreign table
	Alias string // alias for the foreign table

	OurKey     string // key in the main table
	ForeignKey string // key in the other table

	ForeignColumn string // name of the column to print
}

func (l LeftJoin) selectClause(table, name string) string {
	return fmt.Sprintf("%s.%s as %s", l.Alias, l.ForeignColumn, name)
}

func (l LeftJoin) appendClause(table, name string) string {
	return fmt.Sprintf("LEFT JOIN %s AS %s ON %s.%s = %s.%s", l.Table, l.Alias, table, l.OurKey, l.Alias, l.ForeignKey)
}
