package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

func doJSON(pb *pathbuilder.Pathbuilder, index *sparkl.Index, bEngine storages.BundleEngine) {
	var err error

	// generate bundles
	var bundles map[string][]sparkl.Entity
	{
		start := perf.Now()
		bundles, err = sparkl.LoadPathbuilder(pb, index, bEngine)
		if err != nil {
			log.Fatalf("Unable to load pathbuilder: %s", err)
		}
		bundleT := perf.Since(start)
		log.Printf("extracted bundles, took %s", bundleT)
	}

	json.NewEncoder(os.Stdout).Encode(bundles)
}
