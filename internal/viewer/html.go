package viewer

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/internal/htmlx"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

//go:embed templates/*
var templates embed.FS

var parsedTemplates = (func() *template.Template {
	return template.Must(
		template.New("").Funcs(template.FuncMap{
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
		}).ParseFS(
			templates, "templates/*.html", "templates/fragments/*.html",
		),
	)
})()

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
	err := parsedTemplates.ExecuteTemplate(w, "index.html", htmlIndexContext{
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
	err := parsedTemplates.ExecuteTemplate(w, "bundle.html", htmlBundleContext{
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

	bundle, ok := viewer.uri2bundle(vars["uri"])
	if !ok {
		http.NotFound(w, r)
		return
	}

	// redirect to the entity
	target := "/entity/" + bundle + "?uri=" + url.PathEscape(vars["uri"])
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

type htmlEntityContext struct {
	Globals contextGlobal

	Bundle *pathbuilder.Bundle
	Entity *exporter.Entity
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
	err := parsedTemplates.ExecuteTemplate(w, "entity.html", htmlEntityContext{
		Globals: viewer.contextGlobal(),

		Bundle: bundle,
		Entity: entity,
	})
	if err != nil {
		log.Println(err)
	}
}
