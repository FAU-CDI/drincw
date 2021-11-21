package drincw

import (
	"strings"
)

func (pb Pathbuilder) Text() string {
	var builder strings.Builder
	for _, b := range pb.Bundles() {
		b.textIndent(&builder, "", "  ")
	}
	return builder.String()
}

func (bundle Bundle) textIndent(builder *strings.Builder, prefix, indent string) {
	bundle.Group.textIndent(builder, "Bundle", prefix, indent)
	for _, b := range bundle.Bundles() {
		b.textIndent(builder, prefix+indent, indent)
	}
	for _, f := range bundle.Fields() {
		f.textIndent(builder, prefix+indent, indent)
	}
}

func (field Field) textIndent(builder *strings.Builder, prefix, indent string) {
	field.Path.textIndent(builder, "Field", prefix, indent)
}

func (path Path) textIndent(builder *strings.Builder, kind string, prefix, indent string) {
	builder.WriteString(prefix)
	builder.WriteString(path.UUID)
	builder.WriteString(" (")
	if kind != "" {
		builder.WriteString(kind)
		builder.WriteString(" ")
	}
	builder.WriteString(path.ID)
	builder.WriteString(")")
	builder.WriteString("\n")
}
