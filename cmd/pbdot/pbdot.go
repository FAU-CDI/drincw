// Command pbdot turns a pathbuilder into a dot graph
package main

import (
	"flag"
	"fmt"
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

	opts.Prefixes = map[string]string{
		"ecrm": "http://erlangen-crm.org/200717/",
		"sk":   "https://schreibkalender.wisski.agfd.fau.de/ontology/schreibkalender/",
	}
	g := dot.NewDotForBundles(opts, bundles...)
	g.Write(os.Stdout)
}

var nArgs []string
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

	flag.BoolVar(&opts.IsolateChildBundles, "isolate", false, "Isolate sub bundles")
	flag.BoolVar(&opts.CopyChildBundleNodes, "copy", false, "Copy nodes in child bundles (experimental)")
	flag.BoolVar(&opts.BundleUseDisplayNames, "human", false, "Use human names for bundle labels")
	flag.StringVar(&opts.BundleColor, "color", "red", "Color for bundle heads")

	flag.Parse()
	nArgs = flag.Args()
}
