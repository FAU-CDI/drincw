package exporter

import (
	"io"

	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Exporter handles WissKI Entities
type Exporter interface {
	io.Closer

	// Begin signals that count entities will be transmitted for the given bundle
	Begin(bundle *pathbuilder.Bundle, count int64) error

	// Add adds entities for the given bundle
	Add(bundle *pathbuilder.Bundle, entity *wisski.Entity) error

	// End signals that no more entities will be submitted for the given bundle
	End(bundle *pathbuilder.Bundle) error
}
