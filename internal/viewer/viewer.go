package viewer

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Viewer implements a Viewer for entities
type Viewer struct {
	// Pathbuilder and Data to server
	// Should not be changed once a single request has been served.
	Pathbuilder *pathbuilder.Pathbuilder
	Data        map[string][]exporter.Entity

	init sync.Once
	mux  mux.Router

	biLock  sync.RWMutex
	biIndex map[string]map[string]int
}

func (viewer *Viewer) prepare() {
	viewer.init.Do(func() {
		viewer.mux.HandleFunc("/", viewer.htmlIndex)
		viewer.mux.HandleFunc("/bundle/{bundle}", viewer.htmlBundle)
		viewer.mux.HandleFunc("/entity/{bundle}", viewer.htmlEntity).Queries("uri", "{uri:.+}")

		viewer.mux.HandleFunc("/api/v1", viewer.jsonIndex)
		viewer.mux.HandleFunc("/api/v1/bundle/{bundle}", viewer.jsonBundle)
		viewer.mux.HandleFunc("/api/v1/entity/{bundle}", viewer.jsonEntity).Queries("uri", "{uri:.+}")
	})
}

func (viewer *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	viewer.prepare()
	viewer.mux.ServeHTTP(w, r)
}
