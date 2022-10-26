package sparkl

import (
	"log"
	"sync/atomic"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/igraph"
)

type Paths = igraph.Paths[string, any]
type Path = igraph.Path[string, any]

const rdfType = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
const datatypeEmpty = "empty"

const debugLogAllPaths = false // turn this on to log all paths being queried

var debugLogID int64 // id of the current log id

// Path returns the path values of a given path
func FromPath(path pathbuilder.Path, index *Index) []Path {
	// start with the path array
	uris := append([]string{}, path.PathArray...)
	if len(uris) == 0 {
		return nil
	}

	// add the datatype property if are not a group
	// and it is not empty
	if !path.IsGroup && path.DatatypeProperty != "" && path.DatatypeProperty != datatypeEmpty {
		uris = append(uris, path.DatatypeProperty)
	}

	// if debugging is enabled, set it up
	var debugID int64
	if debugLogAllPaths {
		debugID = atomic.AddInt64(&debugLogID, 1)
	}

	set := index.PathsStarting(rdfType, uris[0])
	if debugLogAllPaths {
		log.Println(debugID, uris[0], set.Size())
	}

	for i := 1; i < len(uris) && set.Size() > 0; i++ {
		if i%2 == 0 {
			set.Ending(rdfType, uris[i])
		} else {
			set.Connected(uris[i])
		}

		if debugLogAllPaths {
			log.Println(debugID, uris[i], set.Size())
		}
	}

	return set.Paths()
}
