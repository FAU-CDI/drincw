package viewer

import (
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// findBundle returns a bundle by id and makes sure the caches for the given bundle as filled.
func (viewer *Viewer) findBundle(id string) (bundle *pathbuilder.Bundle, ok bool) {
	bundle = viewer.Pathbuilder.Get(id)
	if bundle == nil {
		return nil, false
	}

	viewer.biLock.Lock()
	defer viewer.biLock.Unlock()

	// fetch the cache for looking up uris for the given bundle
	// if it doesn't exist, make it!
	_, ok = viewer.biIndex[bundle.Group.ID]
	if !ok {
		entities := viewer.Data[bundle.Group.ID]
		index := make(map[string]int, len(entities))
		for i, e := range entities {
			index[e.URI] = i
		}
		if viewer.biIndex == nil {
			viewer.biIndex = make(map[string]map[string]int, len(viewer.Data))
		}
		viewer.biIndex[bundle.Group.ID] = index
	}

	return bundle, true
}

// findEntity finds an entity by the given bundle id
func (viewer *Viewer) findEntity(bundleid, uri string) (bundle *pathbuilder.Bundle, entity *exporter.Entity, ok bool) {
	bundle, ok = viewer.findBundle(bundleid)
	if !ok {
		return nil, nil, false
	}

	viewer.biLock.RLock()
	defer viewer.biLock.RUnlock()

	// find the index of the given URI
	idx, ok := viewer.biIndex[bundle.Group.ID][uri]
	if !ok {
		return nil, nil, false
	}

	// return the entity
	entity = &viewer.Data[bundle.Group.ID][idx]
	ok = true
	return
}

// getBundleNames returns the list of bundles
func (viewer *Viewer) getBundleNames() []string {
	bundles := maps.Keys(viewer.Data)
	slices.Sort(bundles)
	return bundles
}

func (viewer *Viewer) getBundles() (bundles []*pathbuilder.Bundle, ok bool) {
	names := viewer.getBundleNames()
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

	entities := viewer.Data[bundle.Group.ID]
	uris = make([]string, len(entities))
	for i, entity := range entities {
		uris[i] = entity.URI
	}
	return bundle, uris, true
}

// getEntityURIs returns the URIs belonging to a single bundle
// TODO: Make this stream
func (viewer *Viewer) getEntity(id, uri string) (entity *exporter.Entity, ok bool) {
	_, entity, ok = viewer.findEntity(id, uri)
	return
}
