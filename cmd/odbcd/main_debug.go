//go:build debug

package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// enable debug mode
const debugEnabled = true

func init() {
	log.Println("debug mode enabled")
}

//
// DIST FILES
//

func init() {
	_, mainGoPath, _, _ := runtime.Caller(0)
	distPath := filepath.Join(filepath.Dir(mainGoPath), "dist")
	distFS = os.DirFS(distPath)
}
