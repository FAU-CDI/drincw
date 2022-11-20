package sql

import (
	"reflect"
	"testing"
)

func TestMarshalUnmarshalSelector(t *testing.T) {
	tests := []struct {
		name      string
		selector  Selector
		marsheled string
	}{
		{
			name: "simple column selector",
			selector: &ColumnSelector{
				Column: "name",
			},
			marsheled: "`column` name",
		},
		{
			name: "join selector",
			selector: &JoinSelector{
				Column:   "Column",
				Table:    "Table",
				OurKey:   "OurKey",
				TheirKey: "TheirKey",
			},
			marsheled: "`join` `Column` `from` `Table` on OurKey TheirKey",
		},
		{
			name: "many2many selector",
			selector: &Many2ManySelector{
				Column:          "Column",
				Table:           "Table",
				Through:         "Through",
				TheirKey:        "TheirKey",
				TheirThroughKey: "TheirThroughKey",
				OurThroughKey:   "OurThroughKey",
				OurKey:          "OurKey",
			},
			marsheled: "many2many `Column` `from` `Table` through Through on TheirKey TheirThroughKey OurThroughKey OurKey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalSelector(tt.selector)
			if (err != nil) != false {
				t.Errorf("MarshalSelector() error = %v, wantErr %v", err, false)
				return
			}
			if got != tt.marsheled {
				t.Errorf("MarshalSelector() = %v, want %v", got, tt.marsheled)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalSelector(tt.marsheled)
			if (err != nil) != false {
				t.Errorf("UnmarshalSelector() error = %v, wantErr %v", err, false)
				return
			}
			if !reflect.DeepEqual(got, tt.selector) {
				t.Errorf("UnmarshalSelector() = %v, want %v", got, tt.selector)
			}
		})
	}
}
