package viewer

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

//go:embed templates/*
var templates embed.FS

var parsedTemplates = (func() *template.Template {
	return template.Must(
		template.New("").Funcs(template.FuncMap{
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

func (viewer *Viewer) htmlIndex(w http.ResponseWriter, r *http.Request) {
	bundles, ok := viewer.getBundles()
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	err := parsedTemplates.ExecuteTemplate(w, "index.html", bundles)
	if err != nil {
		panic(err)
	}
}

type htmlBundleContext struct {
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
		Bundle: bundle,
		URIS:   entities,
	})
	if err != nil {
		panic(err)
	}
}

type htmlEntityContext struct {
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
		Bundle: bundle,
		Entity: entity,
	})
	if err != nil {
		log.Println(err)
	}
}
