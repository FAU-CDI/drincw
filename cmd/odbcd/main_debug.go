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
// STATIC FILES
//

func init() {
	_, mainGoPath, _, _ := runtime.Caller(0)
	staticPath := filepath.Join(filepath.Dir(mainGoPath), "static")
	staticFS = os.DirFS(staticPath)
}
