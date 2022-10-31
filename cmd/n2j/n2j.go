// Command n2r turns an nquads file into a json file
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

func main() {
	if len(nArgs) != 2 {
		log.Print("Usage: n2r [-help] [...flags] /path/to/pathbuilder /path/to/nquads")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error

	// read the pathbuilder
	var pb pathbuilder.Pathbuilder
	{
		start := perf.Now()
		pb, err = pbxml.Load(nArgs[0])
		pbT := perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to load Pathbuilder: %s", err)
		}
		log.Printf("loaded pathbuilder, took %s", pbT)
	}

	var predicates sparkl.Predicates
	sparkl.ParsePredicateString(&predicates.SameAs, sameAs)
	sparkl.ParsePredicateString(&predicates.InverseOf, inverseOf)

	// make an engine
	engine := sparkl.NewEngine(cache)
	bEngine := storages.NewBundleEngine(cache)

	if cache != "" {
		log.Printf("caching data on-disk at %s", cache)
	}

	// build an index
	var index *sparkl.Index
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nArgs[1], predicates, engine)
		indexT := perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		defer index.Close()

		count, err := index.TripleCount()
		if err != nil {
			log.Fatalf("Unable to get triple count: %s", err)
		}
		log.Printf("built index, size %d, took %s", count, indexT)
	}

	switch {
	case sqlite != "":
		doSqlite(&pb, index, bEngine)
	default:
		doJSON(&pb, index, bEngine)
	}
}

// ===================

var nArgs []string
var cache string
var sameAs = string(wisski.SameAs)
var inverseOf = string(wisski.InverseOf)
var sqlite string

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")

	flag.StringVar(&sameAs, "sameas", sameAs, "SameAs Properties")
	flag.StringVar(&inverseOf, "inverseof", inverseOf, "InverseOf Properties")

	flag.StringVar(&cache, "cache", cache, "During indexing, cache data in the given directory as opposed to memory")
	flag.StringVar(&sqlite, "sqlite", sqlite, "Export an sqlite database to the given path")

	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.Parse()
	nArgs = flag.Args()
}
