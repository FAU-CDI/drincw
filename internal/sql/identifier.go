package sql

// cspell:words twiesing

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

// Identifier represents an SQL Identifier
//
// An identifier can be quoted or escaped, see the Quote() and Escape() methods.
// Both operations make the identifier safe to use directly in SQL strings.
//
// Identifier implements fmt.Formatter, see Format.
type Identifier string

// Format implements the fmt.Formatter interface
// The 's' verb escapes the identifier, the 'q' verb quotes the identifier.
func (identifier Identifier) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		f.Write([]byte(identifier.Escaped()))
	case 'q':
		f.Write([]byte(identifier.Quoted()))
	default:
		fmt.Fprintf(f, "%"+string(verb), string(identifier))
	}
}

// Quoted is like Quote, but returns only the first value
func (identifier Identifier) Quoted() string {
	value, _ := identifier.Quote()
	return value
}

// Escaped is is like Escape, but only returns the first value
func (identifier Identifier) Escaped() string {
	value, _ := identifier.Escape()
	return value
}

// Escape escapes this identifier into a string safe for usage within a MariaDB query.
// Escape performs quoting of the identifier only if necessary.
//
// If value is not a valid identifier (neither quoted nor unquoted), returns it unchanged and ok=false.
func (identifier Identifier) Escape() (escaped string, ok bool) {
	valid, needsQuote, count := identifier.check()
	if !valid {
		return string(identifier), false
	}

	if !needsQuote {
		return string(identifier), true
	}

	return identifier.quote(count), true
}

func (identifier Identifier) Quote() (quoted string, ok bool) {
	valid, _, count := identifier.check()
	if !valid {
		return string(identifier), false
	}
	return identifier.quote(count), true
}

var builderPool = &sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

var RUNE_QUOTE = '`'

// quote quotes an identifier without performing any internal checks.
//
// guess should contain the number of quotes inside the identifier.
// It is used only for optimization purposes
func (identifier Identifier) quote(guess int) string {
	// grab a new builder from the pool
	builder := builderPool.Get().(*strings.Builder)
	builder.Reset()
	defer builderPool.Put(builder)

	builder.Grow(len(identifier) + 2 + guess)

	// iterate over the builder, and quote only the "`" character
	builder.WriteRune(RUNE_QUOTE)
	for _, r := range identifier {
		if r == RUNE_QUOTE {
			builder.WriteRune(RUNE_QUOTE)
		}
		builder.WriteRune(r)
	}
	builder.WriteRune(RUNE_QUOTE)

	return builder.String()
}

// restrictedKeywords contains a list of sql keywords
var restrictedKeywords = map[string]struct{}{
	"add": {}, "all": {}, "alter": {}, "and": {}, "any": {}, "as": {}, "asc": {}, "avg": {}, "backup": {}, "between": {}, "by": {}, "case": {}, "check": {}, "column": {}, "constraint": {}, "count": {}, "create": {}, "database": {}, "default": {}, "delete": {}, "desc": {}, "distinct": {}, "drop": {}, "exec": {}, "exists": {}, "foreign": {}, "from": {}, "full": {}, "group": {}, "having": {}, "in": {}, "index": {}, "inner": {}, "insert": {}, "into": {}, "is": {}, "join": {}, "key": {}, "left": {}, "like": {}, "limit": {}, "max": {}, "min": {}, "not": {}, "null": {}, "or": {}, "order": {}, "outer": {}, "primary": {}, "procedure": {}, "replace": {}, "right": {}, "rownum": {}, "select": {}, "set": {}, "sql": {}, "sum": {}, "table": {}, "top": {}, "truncate": {}, "union": {}, "unique": {}, "update": {}, "values": {}, "view": {}, "where": {},
}

// check checks if an identifier is valid
// valid indicates if the identifier is valid at all.
// needsQuote indicates if the identifier needs to be quoted.
// quoteCharCount indicates the number of characters that need to be prefixed with a quote character.
//
// Adapted from https://mariadb.com/kb/en/identifier-names/#quote-character.
func (identifier Identifier) check() (valid bool, needsQuote bool, quoteCharCount int) {
	// an identifier may not be empty
	if len(identifier) == 0 {
		return false, false, 0
	}

	var lastRune rune     // the last rune in the string
	sawOnlyDigits := true // does the identifier contain only digits?
	for _, r := range identifier {
		// must be part of an identifier
		if !('\u0001' <= r && r <= '\uffff') {
			return false, false, 0
		}

		// identifier starting with digits followed by 'e' must be escaped
		// to prevent confusion with a literal
		if sawOnlyDigits && (r == 'e') {
			needsQuote = true
		}

		// did we see a digit?
		isDigit := '0' <= r && r <= '9'
		if !isDigit {
			sawOnlyDigits = false
		}

		// characters only allowed in quoted identifiers
		if !(isDigit || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || (r == '$') || (r == '_') || ('\u0080' <= r && r <= '\uffff')) {
			needsQuote = true
			if r == RUNE_QUOTE {
				quoteCharCount++
			}
		}

		lastRune = r
	}

	// an identifier may not consist of only digits
	// unless it is quoted
	if sawOnlyDigits {
		return true, true, 0
	}

	// an identifier may not end with a space character
	if unicode.IsSpace(lastRune) {
		return false, false, 0
	}

	// check for restricted keywords (which aren't already cloned)
	if !needsQuote {
		_, ok := restrictedKeywords[strings.ToLower(string(identifier))]
		if ok {
			needsQuote = true
		}
	}

	return true, needsQuote, quoteCharCount
}

// TokenizeIdentifiers is like strings.Split, except that instead of splitting only by spaces
// uses GobbleIdentifiers instead
func TokenizeIdentifiers(value string) (results []Identifier) {
	// NOTE(twiesing): This function is untested because gobbleIdentifier is tested
	var identifier Identifier
	for {
		identifier, value = gobbleIdentifier(value)
		if identifier == "" {
			break
		}
		results = append(results, identifier)
	}
	return
}

// gobbleIdentifier gobbles a single identifier from value, and returns the remaining text of the string.
// If no identifiers are left in value, returns an empty identifier.
func gobbleIdentifier(value string) (identifier Identifier, rest string) {
	// trim spacing
	value = strings.TrimLeftFunc(value, unicode.IsSpace)
	if len(value) == 0 { // nothing left!
		return "", ""
	}

	// fast path: value does not start with the quote rune
	if value[0] != byte(RUNE_QUOTE) {
		return gobbleIdentifierFast(value)
	}

	// slow path: need to work with quotes
	return gobbleIdentifierSlow(value)
}

// gobbleIdentifierFast gobbles an identifier from the start of value.
// value must not start with RUNE_QUOTE.
func gobbleIdentifierFast(value string) (identifier Identifier, rest string) {
	index := len(value)
	for i, r := range value {
		if unicode.IsSpace(r) || r == RUNE_QUOTE {
			index = i
			break
		}
	}
	return Identifier(value[:index]), value[index:]
}

// gobbleIdentifierFast gobbles an identifier from the start of value.
// value must start with RUNE_QUOTE.
func gobbleIdentifierSlow(value string) (identifier Identifier, rest string) {
	doubleQuote := false // did we at any point see a double QUOTE_CHARACTER and need to escape it?
	closed := false

	index := len(value) - 1
	{
		sawQuote := false // did we see a quote in the last index?
		for i, r := range value[1:] {
			if sawQuote {
				// quote followed by non-quote => we are closed
				// so we can break out of the entire loop
				if r != RUNE_QUOTE {
					index = i
					closed = true
					break
				}

				// double quote character => we continue
				sawQuote = false
				doubleQuote = true // we did see a double quote character
				closed = false

				continue
			}

			// saw a quote character
			if r == RUNE_QUOTE {
				sawQuote = true
				closed = true
				continue
			}

		}
	}

	// the rest is whatever is left from the string
	rest = value[index+1:]

	// if the string was not closed, include the last character
	if !closed {
		index++
	}

	// identifier goes up to the index
	found := value[1:index]

	// if we had a double quote, we need to replace them in the input
	if doubleQuote {
		found = strings.ReplaceAll(found, string(RUNE_QUOTE)+string(RUNE_QUOTE), string(RUNE_QUOTE))
	}

	return Identifier(found), rest
}

// IdentifierFactory can generate identifiers with a prefix
type IdentifierFactory Identifier

// Get gets an identifier from this factory
func (idf IdentifierFactory) Get(value string) Identifier {
	if value == "" {
		return Identifier(string(idf))
	}
	return Identifier(string(idf) + "_" + value)
}
