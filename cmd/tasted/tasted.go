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
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/viewer"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
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

	if sameAs != "" {
		flags.SameAsPredicates = strings.Split(sameAs, ",")
	}
	if inverseOf != "" {
		flags.InverseOfPredicates = strings.Split(inverseOf, ",")
	}

	// build an index
	var index *sparkl.Index
	var indexPerf perf.Diff
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nArgs[1], flags.SameAsPredicates, flags.InverseOfPredicates)
		indexPerf = perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		log.Printf("built index, size %d, took %s", index.TripleCount(), indexPerf)
	}

	// generate bundles
	var bundles map[string][]sparkl.Entity
	var bundlesPerf perf.Diff
	{
		start := perf.Now()
		bundles = sparkl.LoadPathbuilder(&pb, index)
		bundlesPerf = perf.Since(start)
		log.Printf("extracted bundles, took %s", bundlesPerf)
	}

	// generate cache
	var cache sparkl.Cache
	var cachePerf perf.Diff
	{
		start := perf.Now()
		cache = sparkl.NewCache(bundles, index.IdentityMap())
		cachePerf = perf.Since(start)
		cachePerf.Bytes += indexPerf.Bytes // because the index is now deallocated
		log.Printf("built cache, took %s", cachePerf)
	}

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
var sameAs string = sparkl.SameAs
var inverseOf string = sparkl.InverseOf

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
