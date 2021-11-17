// Command pbfmt formats a pathbuilder
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw"
)

func main() {
	if len(nArgs) != 1 {
		log.Print("Usage: parsepb [-help] [...flags] /path/to/pathbuilder")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, err := open(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to open pathbuilder: %s", err)
	}

	var x drincw.XMLPathbuilder
	if err := xml.Unmarshal(content, &x); err != nil {
		log.Fatalf("Unable to Unmarshal Pathbuilder: %s", err)

	}
	pb := x.Pathbuilder()

	switch {
	case flagAscii: // format as text
		fmt.Println(pb.Text())
	case flagPretty: // format as pretty xml
		bytes, err := xml.MarshalIndent(pb.XML(), "", "    ")
		if err != nil {
			log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
		}
		fmt.Println(string(bytes))
	default: // format as unpretty xml
		bytes, err := xml.Marshal(pb.XML())
		if err != nil {
			log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
		}
		fmt.Println(string(bytes))
	}
}

// open opens a URL or a file
func open(path string) ([]byte, error) {
	if !(strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")) {
		return os.ReadFile(path)
	}

	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

var nArgs []string

var flagAscii bool = false
var flagPretty bool = false

func init() {
	flag.BoolVar(&flagAscii, "ascii", flagAscii, "format as text instead of xml")
	flag.BoolVar(&flagPretty, "pretty", flagPretty, "format as prettified xml")

	flag.Parse()
	nArgs = flag.Args()
}
