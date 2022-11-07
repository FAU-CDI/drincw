package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/dustin/go-humanize"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/internal/viewer"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

const tastedVersion = 1

// Glass represents a stand-alone representation of a WissKI
type Glass struct {
	Version int

	pathbuilder pathbuilder.Pathbuilder
	PBXML       []byte // Pathbuilder holds the xml-serialized form of the pathbuilder

	Flags viewer.RenderFlags

	Cache *sparkl.Cache
}

func Create(pathbuilderPath string, nquadsPath string, cacheDir string, flags viewer.RenderFlags) (glass Glass, err error) {
	// read the pathbuilder
	var pbPerf perf.Diff
	{
		start := perf.Now()
		glass.pathbuilder, err = pbxml.Load(pathbuilderPath)
		pbPerf = perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to load Pathbuilder: %s", err)
			return glass, err
		}
		log.Printf("loaded pathbuilder, took %s", pbPerf)
	}

	// make an engine
	engine := sparkl.NewEngine(cacheDir)
	bEngine := storages.NewBundleEngine(cacheDir)
	if cacheDir != "" {
		log.Printf("caching data on-disk at %s", cacheDir)
	}

	// build an index
	var index *sparkl.Index
	var indexPerf perf.Diff
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nquadsPath, flags.Predicates, engine)
		indexPerf = perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
			return glass, err
		}
		defer index.Close()

		log.Printf("built index, stats %s, took %s", index.Stats(), indexPerf)
	}

	// generate bundles
	var bundles map[string][]sparkl.Entity
	var bundlesPerf perf.Diff
	{
		start := perf.Now()
		bundles, err = sparkl.LoadPathbuilder(&glass.pathbuilder, index, bEngine)
		if err != nil {
			log.Fatalf("Unable to load pathbuilder: %s", err)
		}
		bundlesPerf = perf.Since(start)
		log.Printf("extracted bundles, took %s", bundlesPerf)
	}

	// generate cache
	var cachePerf perf.Diff
	{
		start := perf.Now()

		identities := make(imap.MemoryStorage[sparkl.URI, sparkl.URI])
		index.IdentityMap(&identities)

		cache, err := sparkl.NewCache(bundles, identities)
		if err != nil {
			log.Fatalf("unable to build cache: %s", err)
		}
		glass.Cache = &cache

		cachePerf = perf.Since(start)
		log.Printf("built cache, took %s", cachePerf)
	}

	index.Close()        // We close the index early, because it's no longer needed
	debug.FreeOSMemory() // force returning memory to the os

	glass.Flags = flags
	return glass, nil
}

// Export writes a glass to disk
func Export(path string, glass Glass) (err error) {
	glass.Version = tastedVersion

	{
		start := perf.Now()
		glass.PBXML, err = pbxml.Marshal(glass.pathbuilder)
		if err != nil {
			log.Fatalf("Unable to create export: %s", err)
			return err
		}
		log.Printf("serialized pathbuilder, took %s", perf.Since(start))
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to create export: %s", err)
		return err
	}
	defer f.Close()

	{
		start := perf.Now()

		counter := &ProgressWriter{
			Writer:   f,
			Progress: os.Stderr,
		}

		err = gob.NewEncoder(counter).Encode(glass)
		os.Stderr.WriteString("\r")
		if err != nil {
			log.Fatalf("Unable to encode export: %s", err)
		}
		log.Printf("wrote export, took %s", perf.Since(start).SetBytes(counter.Bytes))
	}

	return err
}

var errInvalidVersion = errors.New("Glass Export: Invalid version")

// Import loads a glass from disk
func Import(path string) (glass Glass, err error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Unable to open export: %s", err)
		return
	}
	defer f.Close()

	// decode the tasted struct
	err = gob.NewDecoder(f).Decode(&glass)
	if err != nil {
		log.Fatalf("Unable to decode export: %s", err)
		return
	}

	// check that the version is correct
	if glass.Version != tastedVersion {
		log.Fatalf("Unable to decode export: %s", errInvalidVersion)
		return glass, errInvalidVersion
	}

	// decode the xml again
	glass.pathbuilder, err = pbxml.Unmarshal(glass.PBXML)
	if err != nil {
		log.Fatalf("Unable to unmarshal export: %s", err)
		return glass, err
	}

	return
}

type ProgressWriter struct {
	io.Writer
	Progress io.Writer
	Bytes    int64
}

func (cw *ProgressWriter) Write(bytes []byte) (int, error) {
	cw.Bytes += int64(len(bytes))
	fmt.Fprintf(cw.Progress, "\r Wrote %s", humanize.Bytes(uint64(cw.Bytes)))
	return cw.Writer.Write(bytes)
}
