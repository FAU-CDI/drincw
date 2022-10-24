package viewer

import (
	"embed"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Viewer implements an [http.Handler] that displays WissKI Entities.
type Viewer struct {
	// Pathbuilder and Data to server
	// Should not be changed once a single request has been served.
	Pathbuilder *pathbuilder.Pathbuilder
	Data        map[string][]exporter.Entity

	// SameAs database for URIs
	SameAs map[string]string

	alLock sync.Mutex
	alias  map[string][]string

	RenderFlags RenderFlags

	init sync.Once
	mux  mux.Router

	biLock  sync.RWMutex
	biIndex map[string]map[string]int

	ebLock  sync.Mutex
	ebIndex map[string]string
}

type RenderFlags struct {
	HTMLRender  bool   // should we render "text_long" as actual html?
	ImageRender bool   // should we render "image" as actual images
	PublicURL   string // should we replace links from the provided wisski?

	SameAsPredicates []string // SameAsPredicates displayed
}

//go:embed static
var staticEmbed embed.FS

func (viewer *Viewer) prepare() {
	viewer.init.Do(func() {
		viewer.mux.HandleFunc("/", viewer.htmlIndex)
		viewer.mux.HandleFunc("/bundle/{bundle}", viewer.htmlBundle)
		viewer.mux.HandleFunc("/entity/{bundle}", viewer.htmlEntity).Queries("uri", "{uri:.+}")

		viewer.mux.HandleFunc("/wisski/get", viewer.htmlEntityResolve).Queries("uri", "{uri:.+}")

		viewer.mux.HandleFunc("/api/v1", viewer.jsonIndex)
		viewer.mux.HandleFunc("/api/v1/bundle/{bundle}", viewer.jsonBundle)
		viewer.mux.HandleFunc("/api/v1/entity/{bundle}", viewer.jsonEntity).Queries("uri", "{uri:.+}")

		viewer.mux.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticEmbed)))
	})
}

func (viewer *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	viewer.prepare()
	viewer.mux.ServeHTTP(w, r)
}
