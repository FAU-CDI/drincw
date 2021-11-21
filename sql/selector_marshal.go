package sql

import (
	"errors"
	"fmt"
	"strings"
)

var errLineMissingType = errors.New("UnmarshalSelector: selector missing type")

func UnmarshalSelector(line string) (selector Selector, err error) {
	fields := TokenizeIdentifiers(line)
	if len(fields) <= 1 {
		return nil, errLineMissingType
	}

	switch fields[0] {
	case "column":
		selector = &ColumnSelector{}
	case "join":
		selector = &JoinSelector{}
	default:
		return nil, fmt.Errorf("ParseLine: unknown selector type %q", fields[1])
	}

	err = selector.unmarshalFields(fields[1:])
	if err != nil {
		return nil, err
	}

	return selector, nil
}

func MarshalSelector(selector Selector) (string, error) {
	var fields []string

	switch selector.(type) {
	case *ColumnSelector:
		fields = append(fields, "column")
	default:
		return "", fmt.Errorf("MarshalSelector: unknown selector type")
	}
	fields = append(fields, selector.marshalFields()...)

	for index, field := range fields {
		fields[index], _ = EscapeIdentifier(field)
	}

	return strings.Join(fields, " "), nil
}
