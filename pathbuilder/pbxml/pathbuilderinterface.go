package pbxml

import (
	"encoding/xml"
	"errors"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// this file contains the internal xml pathbuilder implementation
// it is not exposed outside of this package; any calls should go via xml.go

type pathbuilderInterface struct {
	XMLName xml.Name `xml:"pathbuilderinterface"`
	Paths   []path   `xml:"path"`
}

// New creates a new XMLPathbuilder from a pathbuilder
func newPathbuilder(pb pathbuilder.Pathbuilder) (x pathbuilderInterface) {
	paths := pb.Paths()
	x.Paths = make([]path, len(paths))
	for i, p := range paths {
		x.Paths[i] = newPath(p)
	}
	return
}

func (xml pathbuilderInterface) Pathbuilder() pathbuilder.Pathbuilder {
	pb := pathbuilder.NewPathbuilder()
	for _, path := range xml.Paths {
		if !path.Enabled {
			continue
		}

		// get the parent group
		parent := pb.GetOrCreate(string(path.GroupID))

		// if we don't have a group, we have a field!
		if !path.IsGroup {
			if parent == nil { // bundle-less fields shouldn't happen
				continue
			}

			parent.ChildFields = append(parent.ChildFields, pathbuilder.Field{Path: path.Path()})
			continue
		}

		// create a new child group
		group := pb.GetOrCreate(path.ID)
		group.Path = path.Path()
		group.Parent = parent
		if parent != nil {
			parent.ChildBundles = append(parent.ChildBundles, group)
		}
	}
	return pb
}

// xmlPathArray represents a set of paths
type xmlPathArray []string

func (paths xmlPathArray) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	var isX bool
	var startPath xml.StartElement
	for _, path := range paths {
		isX = !isX
		if isX {
			startPath.Name.Local = "x"
		} else {
			startPath.Name.Local = "y"
		}

		if err := encoder.EncodeToken(startPath); err != nil {
			return err
		}

		if err := encoder.EncodeToken(xml.CharData([]byte(path))); err != nil {
			return err
		}

		if err := encoder.EncodeToken(startPath.End()); err != nil {
			return err
		}
	}

	return encoder.EncodeToken(start.End())
}

var errPathArrayNoCharData = errors.New("PathArray: Missing CharData for path")
var errPathArrayNoOpen = errors.New("PathArray: Expected opening <x> or <y>")
var errPathArrayNoClose = errors.New("PathArray: Expected closing </x> or </y>")
var errPathArrayInvalid = errors.New("PathArray: Received invalid token")

func (paths *xmlPathArray) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	nextOpen := "x"
	var expectText bool
	var expectClose bool

	results := make([]string, 0)
readloop:
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		// ignore space only tokens
		if text, isText := token.(xml.CharData); isText && strings.TrimSpace(string(text)) == "" {
			continue
		}

		// ignore comments
		if _, isComment := token.(xml.Comment); isComment {
			continue
		}

		switch {
		case expectText:
			text, isText := token.(xml.CharData)
			if !isText {
				return errPathArrayNoCharData
			}

			results = append(results, string(text))

			expectText = false
			expectClose = true
		case expectClose:
			_, isClose := token.(xml.EndElement)
			if !isClose {
				return errPathArrayNoClose
			}
			expectText = false
			expectClose = false
		default:
			open, isOpen := token.(xml.StartElement)
			if isOpen {
				if open.Name.Local != nextOpen {
					return errPathArrayNoOpen
				}
				if nextOpen == "x" {
					nextOpen = "y"
				} else {
					nextOpen = "x"
				}
				expectText = true
				expectClose = false
				continue
			}

			_, isClose := token.(xml.EndElement)
			if isClose {
				break readloop
			}
			return errPathArrayInvalid
		}
	}

	*paths = xmlPathArray(results)
	return nil
}
