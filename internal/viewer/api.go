package viewer

import (
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
func (viewer *Viewer) findEntity(bundleid, uri string) (bundle *pathbuilder.Bundle, entity *sparkl.Entity, ok bool) {
	bundle, ok = viewer.findBundle(bundleid)
	if !ok {
		return nil, nil, false
	}

	entity, ok = viewer.Cache.Entity(uri, bundle.Group.ID)
	if !ok {
		return nil, nil, false
	}

	return
}

func (viewer *Viewer) getBundles() (bundles []*pathbuilder.Bundle, ok bool) {
	names := viewer.Cache.BundleNames
	bundles = make([]*pathbuilder.Bundle, len(names))
	for i, name := range names {
		bundles[i] = viewer.Pathbuilder.Get(name)
		if bundles[i] == nil {
			return nil, false
		}
	}
	return bundles, true
}

// getEntityURIs returns the URIs belonging to a single bundle
// TODO: Make this stream
func (viewer *Viewer) getEntityURIs(id string) (bundle *pathbuilder.Bundle, uris []string, ok bool) {
	bundle, ok = viewer.findBundle(id)
	if !ok {
		return nil, nil, false
	}

	entities := viewer.Cache.BEIndex[bundle.Group.ID]
	uris = make([]string, len(entities))
	for i, entity := range entities {
		uris[i] = entity.URI
	}
	return bundle, uris, true
}

// getEntityURIs returns the URIs belonging to a single bundle
// TODO: Make this stream
func (viewer *Viewer) getEntity(id, uri string) (entity *sparkl.Entity, ok bool) {
	_, entity, ok = viewer.findEntity(id, uri)
	return
}
