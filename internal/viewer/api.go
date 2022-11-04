package viewer

import (
	"log"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
)

// findBundle returns a bundle by id and makes sure the caches for the given bundle as filled.
func (viewer *Viewer) findBundle(id string) (bundle *pathbuilder.Bundle, ok bool) {
	bundle = viewer.Pathbuilder.Get(id)
	if bundle == nil {
		return nil, false
	}

	return bundle, true
}

// findEntity finds an entity by the given bundle id
func (viewer *Viewer) findEntity(bundleid string, uri sparkl.URI) (bundle *pathbuilder.Bundle, entity *sparkl.Entity, ok bool) {
	bundle, ok = viewer.findBundle(bundleid)
	if !ok {
		return nil, nil, false
	}

	entity, ok = viewer.Cache.Entity(uri, bundle.Path.ID)
	if !ok {
		return nil, nil, false
	}

	return
}

func (viewer *Viewer) getBundles() (bundles []*pathbuilder.Bundle, ok bool) {
	names := viewer.Cache.BundleNames
	bundles = make([]*pathbuilder.Bundle, 0, len(names))
	for _, name := range names {
		bundle := viewer.Pathbuilder.Get(name)
		if bundle == nil {
			log.Println("nil bundle", name)
			continue
		}
		bundles = append(bundles, bundle)
	}
	return bundles, true
}

// getEntityURIs returns the URIs belonging to a single bundle
// TODO: Make this stream
func (viewer *Viewer) getEntityURIs(id string) (bundle *pathbuilder.Bundle, uris []sparkl.URI, ok bool) {
	bundle, ok = viewer.findBundle(id)
	if !ok {
		return nil, nil, false
	}

	entities := viewer.Cache.BEIndex[bundle.Path.ID]
	uris = make([]sparkl.URI, len(entities))
	for i, entity := range entities {
		uris[i] = entity.URI
	}
	return bundle, uris, true
}

// getEntityURIs returns the URIs belonging to a single bundle
// TODO: Make this stream
func (viewer *Viewer) getEntity(id string, uri sparkl.URI) (entity *sparkl.Entity, ok bool) {
	_, entity, ok = viewer.findEntity(id, uri)
	return
}
