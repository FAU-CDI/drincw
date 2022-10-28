// Command tasted implements a very simple WissKI Viewer
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/viewer"
	"github.com/tkw1536/FAU-CDI/drincw/internal/wisski"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/imap"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

func main() {
	if len(nArgs) != 2 {
		log.Print("Usage: tasted [-help] [...flags] /path/to/pathbuilder /path/to/nquads")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// start listening, so that even during loading we are not performing that badly
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening on", addr)

	// read the pathbuilder
	var pb pathbuilder.Pathbuilder
	var pbPerf perf.Diff
	{
		start := perf.Now()
		pb, err = pbxml.Load(nArgs[0])
		pbPerf = perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to load Pathbuilder: %s", err)
		}
		log.Printf("loaded pathbuilder, took %s", pbPerf)
	}

	sparkl.ParsePredicateString(&flags.Predicates.SameAs, sameAs)
	sparkl.ParsePredicateString(&flags.Predicates.InverseOf, inverseOf)

	// build an index
	var index *sparkl.Index
	var indexPerf perf.Diff
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nArgs[1], flags.Predicates, &sparkl.MemoryEngine{})
		indexPerf = perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		defer index.Close()

		count, err := index.TripleCount()
		if err != nil {
			log.Fatalf("Unable to get triple count: %s", err)
		}
		log.Printf("built index, size %d, took %s", count, indexPerf)
	}

	// generate bundles
	var bundles map[string][]sparkl.Entity
	var bundlesPerf perf.Diff
	{
		start := perf.Now()
		bundles, err = sparkl.LoadPathbuilder(&pb, index)
		if err != nil {
			log.Fatalf("Unable to load pathbuilder: %s", err)
		}
		bundlesPerf = perf.Since(start)
		log.Printf("extracted bundles, took %s", bundlesPerf)
	}

	// generate cache
	var cache sparkl.Cache
	var cachePerf perf.Diff
	{
		start := perf.Now()

		identities := make(imap.MemoryStorage[sparkl.URI, sparkl.URI])
		index.IdentityMap(identities)
		cache = sparkl.NewCache(bundles, identities)

		cachePerf = perf.Since(start)
		log.Printf("built cache, took %s", cachePerf)
	}

	index.Close() // We close the index early, because it's no longer needed

	// and finally make a viewer handler
	var handler viewer.Viewer
	var handlerPerf perf.Diff
	{
		start := perf.Now()

		handler = viewer.Viewer{
			Cache:       &cache,
			Pathbuilder: &pb,
			RenderFlags: flags,
		}
		handler.Prepare()
		handlerPerf = perf.Since(start)
		log.Printf("built handler, took %s", handlerPerf)
	}

	log.Println(perf.Now())

	http.Serve(listener, &handler)
}

var nArgs []string

var addr string = ":3000"

var flags viewer.RenderFlags
var sameAs string = string(wisski.SameAs)
var inverseOf string = string(wisski.InverseOf)

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.StringVar(&addr, "addr", addr, "Instead of dumping data as json, start up a server at the given address")
	flag.BoolVar(&flags.ImageRender, "images", flags.ImageRender, "Enable rendering of images")
	flag.BoolVar(&flags.HTMLRender, "html", flags.HTMLRender, "Enable rendering of html")
	flag.StringVar(&flags.PublicURL, "public", flags.PublicURL, "Public URL of the wisski the data comes from")
	flag.StringVar(&sameAs, "sameas", sameAs, "SameAs Properties")
	flag.StringVar(&inverseOf, "inverseof", inverseOf, "InverseOf Properties")

	flag.Parse()
	nArgs = flag.Args()
}
