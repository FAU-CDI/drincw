package drincw

import "encoding/xml"

func (dict BundleDict) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	for _, bundle := range dict.Bundles() {
		if err := encoder.Encode(bundle); err != nil {
			return err
		}
	}

	return encoder.EncodeToken(start.End())
}

func (bundle *Bundle) MarshalXML(encoder *xml.Encoder, _ xml.StartElement) error {

	// skip empty bundles!
	if bundle == nil {
		return nil
	}

	// <bundle id="...">
	start := xml.StartElement{
		Name: xml.Name{Local: "bundle"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "id"},
				Value: bundle.Group.UUID,
			},
		},
	}

	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	if err := encoder.EncodeToken(xml.Comment([]byte(" " + bundle.Group.Name + " "))); err != nil {
		return err
	}

	var fieldXML struct {
		XMLName     xml.Name `xml:"field"`
		UUID        string   `xml:"id,attr"`
		MachineName string   `xml:"fieldname"`
		Name        string   `xml:",comment"`
	}

	for _, field := range bundle.Fields() {
		fieldXML.UUID = field.UUID
		fieldXML.MachineName = field.ID
		fieldXML.Name = field.Name
		if err := encoder.Encode(fieldXML); err != nil {
			return err
		}
	}

	for _, bundle := range bundle.Bundles() {
		if err := encoder.Encode(bundle); err != nil {
			return err
		}
	}

	// </bundle>
	return encoder.EncodeToken(start.End())
}
