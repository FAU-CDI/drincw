package sql

import (
	"strings"
	"sync"
	"unicode"
)

// TokenizeIdentifiers is like strings.Split, except that instead of splitting only by spaces
// uses GobbleIdentifiers instead
func TokenizeIdentifiers(value string) (results []string) {
	var identifier string
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
func gobbleIdentifier(value string) (identifier string, rest string) {

	// trim spacing
	value = strings.TrimLeftFunc(value, unicode.IsSpace)
	if len(value) == 0 { // nothing left!
		return "", ""
	}

	// there is no `, so gobble until the next space character
	if value[0] != byte(RUNE_QUOTE) {
		index := 0
		for _, r := range value {
			if unicode.IsSpace(r) {
				break
			}
			index++
		}
		return value[:index], value[index:]
	}

	// if there is a ` scan until the next ` not also followed by a `
	// grab a new builder from the pool
	builder := builderPool.Get().(*strings.Builder)
	builder.Reset()
	defer builderPool.Put(builder)

	index := 1 // first character is a quote by construction
	sawQuote := false
	for _, r := range value[1:] {
		// quote followed by non-quote
		if sawQuote {
			// rune followed by non-rune, so we are done
			if r != RUNE_QUOTE {
				break
			}

			// quote followed by a quote; continue
			sawQuote = false
			builder.WriteRune(RUNE_QUOTE)
			index++
			continue
		}

		index++
		if r == RUNE_QUOTE {
			sawQuote = true
			continue
		}
		builder.WriteRune(r)
	}

	return builder.String(), value[index:]
}

// EscapeIdentifier quotes value if neccessary to be used as an identifer.
// When not needed to use an identifier, returns it unchanged.
//
// If value is not a valid identifier (neither quoted nor unquoted), returns it unchanged and ok=false.
func EscapeIdentifier(value string) (quoted string, ok bool) {
	valid, needsQuote, count := checkIdentifier(value)
	if !valid {
		return quoted, false
	}

	if !needsQuote {
		return value, true
	}

	return quoteInternal(value, count), true
}

// QuoteIdentifier quotes value, regardless if quoting is necessary or not.
//
// When not a valid identifier, returns false.
func QuoteIdentifier(value string) (quoted string, ok bool) {
	valid, _, count := checkIdentifier(value)
	if !valid {
		return "", false
	}

	return quoteInternal(value, count), true
}

var builderPool = &sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

var RUNE_QUOTE = '`'

// quoteInteral quotes an identifier, without performing any checks.
// sizeGuess should be the number of characters needs quoting; it's only used as a performance optimization
func quoteInternal(value string, sizeGuess int) string {
	// grab a new builder from the pool
	builder := builderPool.Get().(*strings.Builder)
	builder.Reset()
	defer builderPool.Put(builder)

	builder.Grow(len(value) + 2 + sizeGuess)

	// iterate over the builder, and quote only the "`" character
	builder.WriteRune(RUNE_QUOTE)
	for _, r := range value {
		if r == RUNE_QUOTE {
			builder.WriteRune(RUNE_QUOTE)
		}
		builder.WriteRune(r)
	}
	builder.WriteRune(RUNE_QUOTE)

	return builder.String()
}

// checkIdentifier checks if an identifier is valid.
// valid indicates if the identifier is valid at all.
// needsQuote indicates if the identifier needs to be quoted.
// quoteCharCount indiciates the number of characters that need to be prefixed with a quote character.
//
// Adapted from https://mariadb.com/kb/en/identifier-names/#quote-character.
func checkIdentifier(identifier string) (valid bool, needsQuote bool, quoteCharCount int) {
	// an identifier may not be empty
	if len(identifier) == 0 {
		return false, false, 0
	}

	var lastRune rune     // the last rune in the string
	sawOnlyDigits := true // does the identifier contain only digits?
	for _, r := range identifier {
		// mmust be part of an identifier
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

	return true, needsQuote, quoteCharCount
}
