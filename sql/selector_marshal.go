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

// allSelectors contains a list of all selector types
var allSelectors = []withName{
	JoinSelector{},
	ColumnSelector{},
}

// newSelector creates a new selector of the provided name
func newSelector(typ Identifier) (Selector, error) {
	var theSelector withName
	for _, s := range allSelectors {
		if s.name() == typ {
			theSelector = s
		}
	}
	if theSelector == nil {
		return nil, fmt.Errorf("unknown selector type %q", typ)
	}

	selector := reflect.New(reflect.TypeOf(theSelector)).Interface()
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
