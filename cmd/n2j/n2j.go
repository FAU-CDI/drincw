// Command n2r turns an nquads file into a json file
package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
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

	// read SameAsPredicates
	var sameAsFlags []string
	if sameAs != "" {
		sameAsFlags = strings.Split(sameAs, ",")
	}

	// read InverseOfPredicates
	var inverseOfFlags []string
	if sameAs != "" {
		inverseOfFlags = strings.Split(inverseOf, ",")
	}

	// build an index
	var index *sparkl.Index
	{
		start := perf.Now()
		index, err = sparkl.LoadIndex(nArgs[1], sameAsFlags, inverseOfFlags)
		indexT := perf.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		log.Printf("built index, size %d, took %s", index.TripleCount(), indexT)
	}

	// generate bundles
	var bundles map[string][]sparkl.Entity
	{
		start := perf.Now()
		bundles = sparkl.LoadPathbuilder(&pb, index)
		bundleT := perf.Since(start)
		log.Printf("extracted bundles, took %s", bundleT)
	}

	// dump as json
	json.NewEncoder(os.Stdout).Encode(bundles)
}

var nArgs []string
var sameAs string = sparkl.SameAs
var inverseOf string = sparkl.InverseOf

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
