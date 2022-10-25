package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "embed"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/internal/assets"
	"github.com/tkw1536/FAU-CDI/drincw/odbc"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/sql"
)

func main() {

	http.HandleFunc("/", indexTemplate)
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(robotsTXT)
	})
	http.Handle("/assets/", assets.AssetHandler)

	http.HandleFunc("/api/v1/makeodbc", func(w http.ResponseWriter, r *http.Request) {
		if isNotPost(w, r) {
			return
		}

		// read the body from the request
		content, err := io.ReadAll(r.Body)
		if isError(err, w, "unable to read request body") {
			return
		}

		// unmarshal some xml
		pb, err := pbxml.Unmarshal(content)
		if isError(err, w, "unable to parse pathbuilder") {
			return
		}

		// create the odbc
		builder := sql.NewBuilder(pb)
		odbcs := odbc.MakeServer(pb)
		err = builder.Apply(&odbcs)
		if isError(err, w, "") {
			return
		}

		// and marshal it!
		w.Header().Set("Content-Type", "text/xml")
		enc := xml.NewEncoder(w)
		enc.Indent("", "    ")
		enc.Encode(odbcs)
	})

	log.Printf("Listening on %s", flagListen)

	http.ListenAndServe(flagListen, nil)
}

func isNotPost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return false
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("method not allowed"))
	return true
}

func isError(err error, w http.ResponseWriter, userMessage string) bool {
	if err == nil {
		return false
	}

	if userMessage != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(userMessage))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}
	return true
}

//
// STATIC FILES
//

//go:embed robots.txt
var robotsTXT []byte

//go:embed index.html
var indexHTML string

var indexTemplate = assets.Assetsodbc.MustMakeFunc(
	"index.html", indexHTML,
	template.FuncMap{},
	nil,
)

//
// COMMAND LINE FLAGS
//

var flagListen string = "localhost:8080"

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.StringVar(&flagListen, "listen", flagListen, "address to listen on")

	flag.Parse()
}
