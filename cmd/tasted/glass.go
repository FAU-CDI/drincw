package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/internal/viewer"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/sgob"
)

const tastedVersion = 1

// Glass represents a stand-alone representation of a WissKI
type Glass struct {
	Pathbuilder pathbuilder.Pathbuilder

	Flags viewer.RenderFlags

	Cache *sparkl.Cache
}

func (glass *Glass) EncodeTo(encoder *gob.Encoder) error {
	// encode the pathbuilder as xml
	pbxml, err := pbxml.Marshal(glass.Pathbuilder)
	if err != nil {
		return err
	}

	// encode all the fields
	for _, obj := range []any{
		tastedVersion,
		pbxml,
		glass.Flags,
	} {
		if err := sgob.Encode(encoder, obj); err != nil {
			return err
		}
	}

	// encode the paypload
	return glass.Cache.EncodeTo(encoder)
}

func (glass *Glass) DecodeFrom(decoder *gob.Decoder) (err error) {
	var version int
	var xml []byte
	for _, obj := range []any{
		&version,
		&xml,
		&glass.Flags,
	} {
		if err := sgob.Decode(decoder, obj); err != nil {
			return err
		}
	}

	// decode the xml again
	glass.Pathbuilder, err = pbxml.Unmarshal(xml)
	if err != nil {
		log.Fatalf("Unable to unmarshal export: %s", err)
		return err
	}

	if version != tastedVersion {
		return errInvalidVersion
	}

	glass.Cache = new(sparkl.Cache)
	return glass.Cache.DecodeFrom(decoder)
}

func Create(pathbuilderPath string, nquadsPath string, cacheDir string, flags viewer.RenderFlags) (glass Glass, err error) {
	// read the pathbuilder
	var pbPerf perf.Diff
	{
		start := perf.Now()
		glass.Pathbuilder, err = pbxml.Load(pathbuilderPath)
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
		bundles, err = sparkl.LoadPathbuilder(&glass.Pathbuilder, index, bEngine)
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
		err = glass.EncodeTo(gob.NewEncoder(counter))
		counter.Flush(true)
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
	defer debug.FreeOSMemory() // force clearing free memory

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Unable to open export: %s", err)
		return
	}
	defer f.Close()

	{
		start := perf.Now()

		counter := &ProgressReader{
			Reader:   f,
			Progress: os.Stderr,
		}
		err = glass.DecodeFrom(gob.NewDecoder(counter))
		counter.Flush(true)
		os.Stderr.WriteString("\r")
		if err != nil {
			log.Fatalf("Unable to decode export: %s", err)
		}
		log.Printf("read export, took %s", perf.Since(start).SetBytes(counter.Bytes))
	}

	return
}

type ProgressWriter struct {
	io.Writer
	Bytes int64

	lastFlush time.Time
	Progress  io.Writer
}

func (cw *ProgressWriter) Write(bytes []byte) (int, error) {
	cw.Bytes += int64(len(bytes))
	fmt.Fprintf(cw.Progress, "\r Wrote %s", humanize.Bytes(uint64(cw.Bytes)))
	return cw.Writer.Write(bytes)
}

func (cw *ProgressWriter) Flush(force bool) {
	if force || time.Since(cw.lastFlush) > flushInterval {
		cw.lastFlush = time.Now()
		fmt.Fprintf(cw.Progress, "\r Read %s", humanize.Bytes(uint64(cw.Bytes)))
	}
}

type ProgressReader struct {
	io.Reader
	Bytes int64

	lastFlush time.Time
	Progress  io.Writer
}

func (cr *ProgressReader) Read(bytes []byte) (int, error) {
	count, err := cr.Reader.Read(bytes)
	cr.Bytes += int64(count)
	cr.Flush(false)
	return count, err
}

const flushInterval = time.Second / 30

func (cr *ProgressReader) Flush(force bool) {
	if force || time.Since(cr.lastFlush) > flushInterval {
		cr.lastFlush = time.Now()
		fmt.Fprintf(cr.Progress, "\r Read %s", humanize.Bytes(uint64(cr.Bytes)))
	}
}
