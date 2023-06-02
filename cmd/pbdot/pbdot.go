// Command pbdot turns a pathbuilder into a dot graph
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/FAU-CDI/drincw"
	"github.com/FAU-CDI/drincw/pathbuilder"
	"github.com/FAU-CDI/drincw/pathbuilder/dot"
	"github.com/FAU-CDI/drincw/pathbuilder/pbxml"
	"golang.org/x/exp/maps"
)

func main() {
	if len(nArgs) < 1 {
		log.Print("Usage: pbdot [-help] [...flags] /path/to/pathbuilder bundles...")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := loadPrefixMap(prefixMap); err != nil {
		log.Fatal(err)
	}

	pb, err := pbxml.Load(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to load Pathbuilder: %s", err)
	}

	bundles := pb.Bundles()
	if len(nArgs) > 1 {
		bm := make(map[string]*pathbuilder.Bundle)
		for _, b := range nArgs[1:] {
			bm[b] = pb.FindBundle(b)
		}
		bundles = maps.Values(bm)
	}

	g := dot.NewDotForBundles(opts, bundles...)
	g.Write(os.Stdout)
}

func loadPrefixMap(path string) error {
	if path == "" {
		return nil
	}

	var f io.Reader
	if path != "-" {
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		f = file
	} else {
		f = os.Stdin
	}

	return json.NewDecoder(f).Decode(&opts.Prefixes)
}

var nArgs []string
var prefixMap string
var opts dot.Options

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.BoolVar(&opts.BundleUseDisplayNames, "human-bundle-names", false, "Use human names for bundle labels")

	flag.BoolVar(&opts.FlatChildBundles, "flat", false, "Skip sub-bundle structure entirely")
	flag.BoolVar(&opts.IndependentChildBundles, "isolate-child-bundles", false, "Render each child bundle independently")
	flag.BoolVar(&opts.CopyChildBundleNodes, "copy-child-bundle-nodes", false, "Copy nodes in child bundles (experimental)")

	flag.StringVar(&opts.ColorBundle, "color-heads", "red", "Color for bundle heads")
	flag.StringVar(&opts.ColorDatatype, "color-data", "blue", "Color for datatypes")

	flag.StringVar(&prefixMap, "prefixes", "", "Load prefixes in json format from the given file")

	flag.Parse()
	nArgs = flag.Args()
}
