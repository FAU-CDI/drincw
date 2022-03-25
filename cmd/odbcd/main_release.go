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

//go:embed dist/*
var embedDistDir embed.FS

func init() {
	var err error
	distFS, err = fs.Sub(embedDistDir, "dist")
	if err != nil {
		panic(err)
	}
}
