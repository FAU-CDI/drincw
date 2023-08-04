// Package pbxml implements the XML formats for a pathbuilder.
package pbxml

// cspell:words pbxml pathbuilder

import (
	"encoding/xml"

	"github.com/FAU-CDI/drincw/internal/source"
	"github.com/FAU-CDI/drincw/pathbuilder"
)

// XMLPathbuilder is an XML representation of a Pathbuilder
// It implements xml.Marshaler and xml.Unmarshaler.
//
// It intentionally does not expose any implementation details, as the format might change in the future.
type XMLPathbuilder struct {
	data pathbuilderInterface
}

func (builder XMLPathbuilder) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(builder.data, start)
}

func (builder *XMLPathbuilder) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return d.DecodeElement(&builder.data, &start)
}

func (builder XMLPathbuilder) Pathbuilder() pathbuilder.Pathbuilder {
	return builder.data.Pathbuilder()
}

// New creates a new XMLPathbuilder from a pathbuilder
func New(pb pathbuilder.Pathbuilder) XMLPathbuilder {
	return XMLPathbuilder{data: newPathbuilder(pb)}
}

// Load loads a pathbuilder in xml from src.
// Source can should be either a local path or a remote 'http://' or 'https://' url; see source.ReadAll.
func Load(src string) (pb pathbuilder.Pathbuilder, err error) {
	bytes, err := source.ReadAll(src)
	if err != nil {
		return pb, err
	}
	return Unmarshal(bytes)
}

// Marshal marshals a pathbuilder as XML
func Marshal(pb pathbuilder.Pathbuilder) ([]byte, error) {
	return xml.Marshal(New(pb).data)
}

// Unmarshal un-marshals a pathbuilder from XML
func Unmarshal(data []byte) (pb pathbuilder.Pathbuilder, err error) {
	var xpb pathbuilderInterface
	if err := xml.Unmarshal(data, &xpb); err != nil {
		return pb, err
	}
	return xpb.Pathbuilder(), nil
}
