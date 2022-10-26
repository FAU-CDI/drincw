package viewer

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/assets"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// Viewer implements an [http.Handler] that displays WissKI Entities.
type Viewer struct {
	Cache       *sparkl.Cache
	Pathbuilder *pathbuilder.Pathbuilder

	RenderFlags RenderFlags

	init sync.Once
	mux  mux.Router
}

type RenderFlags struct {
	HTMLRender  bool   // should we render "text_long" as actual html?
	ImageRender bool   // should we render "image" as actual images
	PublicURL   string // should we replace links from the provided wisski?

	SameAsPredicates []string // SameAsPredicates displayed
}

func (viewer *Viewer) Prepare() {
	viewer.init.Do(func() {
		viewer.mux.HandleFunc("/", viewer.htmlIndex)
		viewer.mux.HandleFunc("/bundle/{bundle}", viewer.htmlBundle)
		viewer.mux.HandleFunc("/entity/{bundle}", viewer.htmlEntity).Queries("uri", "{uri:.+}")

		viewer.mux.HandleFunc("/wisski/get", viewer.htmlEntityResolve).Queries("uri", "{uri:.+}")

		viewer.mux.HandleFunc("/api/v1", viewer.jsonIndex)
		viewer.mux.HandleFunc("/api/v1/bundle/{bundle}", viewer.jsonBundle)
		viewer.mux.HandleFunc("/api/v1/entity/{bundle}", viewer.jsonEntity).Queries("uri", "{uri:.+}")

		viewer.mux.PathPrefix("/assets/").Handler(assets.AssetHandler)
	})
}

func (viewer *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	viewer.Prepare()
	viewer.mux.ServeHTTP(w, r)
}
