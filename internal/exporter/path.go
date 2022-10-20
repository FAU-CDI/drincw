package exporter

import (
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

type Index = sparkl.GraphIndex[string, any]
type Paths = sparkl.Paths[string, any]
type Path = sparkl.Path[string, any]

const rdfType = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

// Path returns the path values of a given path
func FromPath(path pathbuilder.Path, index *Index) []Path {
	// TODO: Figure out when we don't need the Datatype property
	uris := append([]string{}, path.PathArray...)
	if len(uris) == 0 {
		return nil
	}

	if !path.IsGroup {
		uris = append(uris, path.DatatypeProperty)
	}

	set := index.PathsStarting(rdfType, uris[0])
	for i := 1; i < len(uris); i++ {
		if i%2 == 0 {
			set.Ending(rdfType, uris[i])
			continue
		} else {
			set.Connected(uris[i])
		}
	}

	return set.Paths()
}
