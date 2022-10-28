// Command n2r turns an nquads file into a json file
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
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

	// build an index
	var index *sparkl.Index
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nArgs[1], predicates)
		indexT := perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		count, err := index.TripleCount()
		if err != nil {
			log.Fatalf("Unable to get triple count: %s", err)
		}
		log.Printf("built index, size %d, took %s", count, indexT)
	}

	// generate bundles
	var bundles map[string][]sparkl.Entity
	{
		start := perf.Now()
		bundles, err = sparkl.LoadPathbuilder(&pb, index)
		if err != nil {
			log.Fatalf("Unable to load pathbuilder: %s", err)
		}
		bundleT := perf.Since(start)
		log.Printf("extracted bundles, took %s", bundleT)
	}

	// dump as json
	json.NewEncoder(os.Stdout).Encode(bundles)
}

var nArgs []string
var sameAs = string(wisski.SameAs)
var inverseOf = string(wisski.InverseOf)

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	flag.StringVar(&sameAs, "sameas", sameAs, "SameAs Properties")
	flag.StringVar(&inverseOf, "inverseof", inverseOf, "InverseOf Properties")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.Parse()
	nArgs = flag.Args()
}
