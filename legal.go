package drincw

import _ "embed"

//go:generate gogenlicense -m -t 0

//go:embed LICENSE
var License string

// LegalText returns legal text to be included in human-readable output using huelio.
func LegalText() string {
	return `
================================================================================
Drincw - Drink Really Is Not Copying WissKI
================================================================================
` + License + "\n" + Notices
}
