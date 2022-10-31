package main

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/exporter"
	"github.com/tkw1536/FAU-CDI/drincw/internal/sparkl/storages"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pkg/perf"
)

func doSqlite(pb *pathbuilder.Pathbuilder, index *sparkl.Index, bEngine storages.BundleEngine) {
	var err error

	// setup the sqlite
	db, err := sql.Open("sqlite", sqlite)
	if err != nil {
		log.Fatal(err)
	}

	// and do the export
	{
		start := perf.Now()
		err = sparkl.Export(pb, index, bEngine, &exporter.SQL{
			DB: db,
		})
		if err != nil {
			log.Fatalf("Unable to export sql: %s", err)
		}
		bundleT := perf.Since(start)
		log.Printf("wrote bundles, took %s", bundleT)
	}
}
