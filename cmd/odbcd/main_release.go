//go:build !debug

package main

import (
	"embed"
	"io/fs"
)

// disable debugging mode
const debugEnabled = false

//
// STATIC FILES
//

//go:embed static/*
var embedStaticDir embed.FS

func init() {
	var err error
	staticFS, err = fs.Sub(embedStaticDir, "static")
	if err != nil {
		panic(err)
	}
}
