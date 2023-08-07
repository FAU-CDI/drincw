package sql

import (
	"reflect"
	"testing"
)

func Test_Identifier_check(t *testing.T) {
	tests := []struct {
		identifier         string
		wantValid          bool
		wantNeedsQuote     bool
		wantQuoteCharCount int
	}{
		{"", false, false, 0}, // empty identifier

		{"azAzQ_$", true, false, 0}, // nothing has to be quoted

		{"hello world", true, true, 0}, // space needs quoting
		{"hello`world", true, true, 1}, // ` needs quoting

		{"join", true, true, 0}, // keyword needs quoting

		{"10e12", true, true, 0}, // things confused with a literal need quoting

		{"0000", true, true, 0},  // only numerals
		{"000a", true, false, 0}, // not only numerals
	}
	for _, tt := range tests {
		t.Run(tt.identifier, func(t *testing.T) {
			gotValid, gotNeedsQuote, gotQuoteCharCount := Identifier(tt.identifier).check()
			if gotValid != tt.wantValid {
				t.Errorf("checkIdentifier() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
			if gotNeedsQuote != tt.wantNeedsQuote {
				t.Errorf("checkIdentifier() gotNeedsQuote = %v, want %v", gotNeedsQuote, tt.wantNeedsQuote)
			}
			if gotQuoteCharCount != tt.wantQuoteCharCount {
				t.Errorf("checkIdentifier() gotQuoteCharCount = %v, want %v", gotQuoteCharCount, tt.wantQuoteCharCount)
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		wantQuoted string
		wantOk     bool
	}{
		{"", "", false}, // empty identifier

		{"azAzQ_$", "`azAzQ_$`", true}, // nothing has to be quoted

		{"hello world", "`hello world`", true},  // space needs quoting
		{"hello`world", "`hello``world`", true}, // ` needs quoting

		{"10e12", "`10e12`", true}, // things confused with a literal need quoting

		{"0000", "`0000`", true}, // only numerals
		{"000a", "`000a`", true}, // not only numerals
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuoted, gotOk := Identifier(tt.name).Quote()
			if gotQuoted != tt.wantQuoted {
				t.Errorf("QuoteIdentifier() gotQuoted = %v, want %v", gotQuoted, tt.wantQuoted)
			}
			if gotOk != tt.wantOk {
				t.Errorf("QuoteIdentifier() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestIdentifier_Escape(t *testing.T) {
	tests := []struct {
		name       string
		wantQuoted string
		wantOk     bool
	}{
		{"", "", false}, // empty identifier

		{"azAzQ_$", "azAzQ_$", true}, // nothing has to be quoted

		{"hello world", "`hello world`", true},  // space needs quoting
		{"hello`world", "`hello``world`", true}, // ` needs quoting

		{"10e12", "`10e12`", true}, // things confused with a literal need quoting

		{"0000", "`0000`", true}, // only numerals
		{"000a", "000a", true},   // not only numerals
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuoted, gotOk := Identifier(tt.name).Escape()
			if gotQuoted != tt.wantQuoted {
				t.Errorf("EscapeIdentifier() gotQuoted = %v, want %v", gotQuoted, tt.wantQuoted)
			}
			if gotOk != tt.wantOk {
				t.Errorf("EscapeIdentifier() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGobbleIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		wantIdentifier Identifier
		wantRest       string
	}{
		{"", "", ""},         // empty
		{"        ", "", ""}, // only spaces

		{"hello", "hello", ""},             // unquoted word
		{"hello world", "hello", " world"}, // unquoted word with rest
		{"   hello", "hello", ""},          // unquoted word with space

		{"`hello`", "hello", ""},              // quoted word without escape
		{"`hello``world`", "hello`world", ""}, // quoted word with escape

		{"`hello` next", "hello", " next"},              // quoted word without escape
		{"`hello``world` next", "hello`world", " next"}, // quoted word with escape

		{"`hello", "hello", ""},              // unclosed quote without `s`
		{"`hello``world", "hello`world", ""}, // unclosed quote with ``s
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdentifier, gotRest := gobbleIdentifier(tt.name)
			if gotIdentifier != tt.wantIdentifier {
				t.Errorf("gobbleIdentifier() gotIdentifier = %v, want %v", gotIdentifier, tt.wantIdentifier)
			}
			if gotRest != tt.wantRest {
				t.Errorf("gobbleIdentifier() gotRest = %v, want %v", gotRest, tt.wantRest)
			}
		})
	}
}

func TestTokenizeIdentifiers(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name        string
		args        args
		wantResults []Identifier
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResults := TokenizeIdentifiers(tt.args.value); !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("TokenizeIdentifiers() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}
