package sparkl

import (
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type Paths = igraph.Paths[string, any]
type Path = igraph.Path[string, any]

const rdfType = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
const datatypeEmpty = "empty"

// Path returns the path values of a given path
func FromPath(path pathbuilder.Path, index *Index) []Path {
	// TODO: Figure out when we don't need the Datatype property
	uris := append([]string{}, path.PathArray...)
	if len(uris) == 0 {
		return nil
	}

	if !path.IsGroup && path.DatatypeProperty != "" && path.DatatypeProperty != datatypeEmpty {
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
