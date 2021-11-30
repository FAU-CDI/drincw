package sql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var errLineMissingType = errors.New("UnmarshalSelector: selector missing type")

func UnmarshalSelector(line string) (selector Selector, err error) {
	fields := TokenizeIdentifiers(line)
	if len(fields) <= 1 {
		return nil, errLineMissingType
	}

	selector, err = newSelector(fields[0])
	if err != nil {
		return nil, err
	}

	return selector, unmarshalSelectorFields(selector, fields[1:])
}

var selectorTypes map[Identifier]reflect.Type

// prefix to add to marshal comment, this is modified by init()
var MARSHAL_COMMENT_PREFIX = `
/*
	This file contains mappings from bundles and fields to sql columns.
	The JSON syntax is self-explanatory; it supports comments using js syntax.

	Supported Fields:

{{examples}}
*/
`

func init() {
	// all known types of selectors, in a sensible order
	selectors := []Selector{
		(*ColumnSelector)(nil),
		(*JoinSelector)(nil),
		(*Many2ManySelector)(nil),
	}

	//
	//

	examples := make([]string, 0, len(selectors))

	selectorTypes = make(map[Identifier]reflect.Type, len(selectors))

	for _, selector := range selectors {
		name := selector.name()
		fields := selector.fields()

		// store the type of the selector in the types map
		selectorTypes[name] = reflect.TypeOf(selector).Elem()

		// generate an example string (by using fields())
		fields = append(fields, "")
		copy(fields[1:], fields[0:])
		fields[0] = string(name)

		for i, f := range fields {
			fields[i] = Identifier(f).Escaped()
		}

		// add the prefix from the string
		examples = append(examples, "\t"+strings.Join(fields, " "))
	}

	MARSHAL_COMMENT_PREFIX = strings.Replace(MARSHAL_COMMENT_PREFIX, "{{examples}}", strings.Join(examples, "\n"), 1)
	MARSHAL_COMMENT_PREFIX = strings.ReplaceAll(MARSHAL_COMMENT_PREFIX, "\t", "    ")
}

// newSelector creates a new selector of the provided name
func newSelector(typ Identifier) (Selector, error) {
	rTyp, ok := selectorTypes[typ]
	if !ok {
		return nil, fmt.Errorf("unknown selector type %q", typ)
	}

	selector := reflect.New(rTyp).Interface()
	return selector.(Selector), nil
}

func unmarshalSelectorFields(dst Selector, src []Identifier) (err error) {
	// recover any panic with an internal errors
	defer func() {
		v := recover()
		if v != nil {
			err = fmt.Errorf("internal error: Selector %q: %s", dst.name(), v)
		}
	}()

	spec := dst.fields()
	if len(spec) != len(src) {
		return fmt.Errorf("Selector %q expected %d arguments, but got %d", dst.name(), len(spec), len(src))
	}

	dstRef := reflect.ValueOf(dst).Elem()

	for i, spec := range spec {
		src := string(src[i])
		if len(spec) == 0 {
			panic("empty field")
		}

		// constant selector: must be equal
		if spec[0] != '$' {
			if src != spec {
				return fmt.Errorf(
					"Selector %q expected %q in position %d, but got %q",
					dst.name(), spec, i, src,
				)
			}
			continue
		}

		// set the field in dst by using the string
		dstRef.FieldByName(spec[1:]).SetString(src)
	}

	return nil
}

func MarshalSelector(selector Selector) (string, error) {
	fields, err := marshalSelectorFields(selector)
	if err != nil {
		return "", err
	}

	// preprend the name of the fields
	fields = append(fields, Identifier(""))
	copy(fields[1:], fields[0:])
	fields[0] = selector.name()

	escaped := make([]string, len(fields))
	for index, field := range fields {
		escaped[index] = field.Escaped()
	}

	return strings.Join(escaped, " "), nil
}

func marshalSelectorFields(src Selector) (identifiers []Identifier, err error) {
	// recover any panic with an internal errors
	defer func() {
		v := recover()
		if v != nil {
			identifiers = nil
			err = fmt.Errorf("internal error: Selector %q: %s", src.name(), v)
		}
	}()

	spec := src.fields()

	srcRef := reflect.ValueOf(src).Elem()

	identifiers = make([]Identifier, len(spec))
	for i, spec := range spec {
		// constant selector: set as equal
		if spec[0] != '$' {
			identifiers[i] = Identifier(spec)
			continue
		}

		// set the field using the provided spec
		identifiers[i] = Identifier(srcRef.FieldByName(spec[1:]).String())
	}

	return identifiers, nil
}
