// Package pbtxt exports a pathbuilder as text
package pbtxt

import (
	"fmt"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Marshal marshals pathbuilder as text
func Marshal(pb pathbuilder.Pathbuilder) string {
	var builder strings.Builder
	for _, b := range pb.Bundles() {
		marshalBundle(&builder, b, "", "  ")
	}
	return builder.String()
}

func marshalBundle(builder *strings.Builder, bundle *pathbuilder.Bundle, prefix, indent string) {
	marshalPath(builder, bundle.Group, "Bundle", prefix, indent)
	for _, b := range bundle.Bundles() {
		marshalBundle(builder, b, prefix+indent, indent)
	}
	for _, f := range bundle.Fields() {
		marshalField(builder, f, prefix+indent, indent)
	}
}

func marshalField(builder *strings.Builder, field pathbuilder.Field, prefix, indent string) {
	marshalPath(builder, field.Path, "Field", prefix, indent)
}

func marshalPath(builder *strings.Builder, path pathbuilder.Path, kind string, prefix, indent string) {
	builder.WriteString(prefix)
	builder.WriteString(path.MachineName())
	builder.WriteString(" (")
	if kind != "" {
		builder.WriteString(kind)
		builder.WriteString(" ")
	}
	builder.WriteString(path.ID)
	builder.WriteString(fmt.Sprintf(" %q", path.Name))
	builder.WriteString(")")
	builder.WriteString("\n")
}
