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
	"time"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
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
		start := time.Now()
		pb, err = pbxml.Load(nArgs[0])
		pbT := time.Since(start)

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

	// build an index
	var index *exporter.Index
	{
		start := time.Now()
		index, err = sparkl.LoadIndex(nArgs[1], sameAsFlags)
		indexT := time.Since(start)

		if err != nil {
			log.Fatalf("Unable to build index: %s", err)
		}
		log.Printf("built index, size %d, took %s", index.TripleCount(), indexT)
	}

	// generate bundles
	var bundles map[string][]exporter.Entity
	{
		start := time.Now()
		bundles = exporter.LoadPathbuilder(&pb, index)
		bundleT := time.Since(start)
		log.Printf("extracted bundles, took %s", bundleT)
	}

	// dump as json
	json.NewEncoder(os.Stdout).Encode(bundles)
}

var nArgs []string
var sameAs string = sparkl.SameAs

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	flag.StringVar(&sameAs, "sameas", sameAs, "SameAs Properties")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.Parse()
	nArgs = flag.Args()
}
