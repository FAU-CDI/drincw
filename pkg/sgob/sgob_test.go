package sgob

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
)

func TestSGob(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "slice",
			value: []string{"hello", "world"},
		},
		{
			name:  "map",
			value: map[string]string{"hello": "world"},
		},
		{
			name:  "map of map",
			value: map[string]map[string]string{"hello": {"hello": "world"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)

			value := reflect.ValueOf(tt.value)

			if err := EncodeValue(encoder, value); err != nil {
				t.Errorf("Encode() error = %v, wantErr %v", err, nil)
			}

			dest := reflect.New(value.Type())

			decoder := gob.NewDecoder(&buffer)
			if err := DecodeValue(decoder, dest); err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, nil)
			}

			got := dest.Elem().Interface()

			if !reflect.DeepEqual(got, tt.value) {
				t.Errorf("sgob round trip got = %v, want %v", got, tt.value)
			}
		})
	}
}
