package viewer

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	_ "embed"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/assets"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/htmlx"
)

var contextTemplateFuncs = template.FuncMap{
	"renderhtml": func(html string, globals contextGlobal) template.HTML {
		return template.HTML(htmlx.ReplaceLinks(html, globals.ReplaceURL))
	},
	"combine": func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, errors.New("pairs must be of even length")
		}
		result := make(map[string]any, len(pairs)/2)
		for i, v := range pairs {
			if i%2 == 1 {
				result[pairs[(i-1)].(string)] = v
			}
		}
		return result, nil
	},
}

//go:embed templates/bundle.html
var bundleHTML string

var bundleTemplate = assets.Assetstasted.MustParseShared(
	"bundle.html",
	bundleHTML,
	contextTemplateFuncs,
)

//go:embed templates/entity.html
var entityHTML string

var entityTemplate = assets.Assetstasted.MustParseShared(
	"entity.html",
	entityHTML,
	contextTemplateFuncs,
)

//go:embed templates/index.html
var indexHTML string

var indexTemplate *template.Template = assets.Assetstasted.MustParseShared(
	"index.html",
	indexHTML,
	contextTemplateFuncs,
)

type contextGlobal struct {
	RenderFlags
	wisskiGetRoute string
}

func (cg contextGlobal) ReplaceURL(url string) string {
	if cg.wisskiGetRoute != "" && strings.HasPrefix(url, cg.wisskiGetRoute) {
		uri := url[len(cg.wisskiGetRoute):]
		return "/wisski/get?uri=" + uri
	}
	return url
}

func (viewer *Viewer) contextGlobal() (global contextGlobal) {
	global.RenderFlags = viewer.RenderFlags

	if viewer.RenderFlags.PublicURL == "" {
		return
	}

	url, err := url.JoinPath(viewer.RenderFlags.PublicURL, "wisski", "get")
	if err != nil {
		return
	}

	global.wisskiGetRoute = url + "?uri="

	return
}

type htmlIndexContext struct {
	Globals contextGlobal
	Bundles []*pathbuilder.Bundle
}

func (viewer *Viewer) htmlIndex(w http.ResponseWriter, r *http.Request) {
	bundles, ok := viewer.getBundles()
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err := indexTemplate.Execute(w, htmlIndexContext{
		Globals: viewer.contextGlobal(),
		Bundles: bundles,
	})
	if err != nil {
		panic(err)
	}
}

type htmlBundleContext struct {
	Globals contextGlobal

	Bundle *pathbuilder.Bundle
	URIS   []string
}

func (viewer *Viewer) htmlBundle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bundle, entities, ok := viewer.getEntityURIs(vars["bundle"])
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err := bundleTemplate.Execute(w, htmlBundleContext{
		Globals: viewer.contextGlobal(),
		Bundle:  bundle,
		URIS:    entities,
	})
	if err != nil {
		panic(err)
	}
}

func (viewer *Viewer) htmlEntityResolve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uri := strings.TrimSpace(vars["uri"])

	bundle, ok := viewer.Cache.Bundle(uri)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// redirect to the entity
	target := "/entity/" + bundle + "?uri=" + url.PathEscape(viewer.Cache.Canonical(uri))
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

type htmlEntityContext struct {
	Globals contextGlobal

	Bundle  *pathbuilder.Bundle
	Entity  *sparkl.Entity
	Aliases []string
}

func (viewer *Viewer) htmlEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bundle, entity, ok := viewer.findEntity(vars["bundle"], vars["uri"])
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err := entityTemplate.Execute(w, htmlEntityContext{
		Globals: viewer.contextGlobal(),

		Bundle:  bundle,
		Entity:  entity,
		Aliases: viewer.Cache.Aliases(entity.URI),
	})
	if err != nil {
		log.Println(err)
	}
}
