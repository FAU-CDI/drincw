// Command pbfmt formats a pathbuilder
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbtxt"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
)

func main() {
	if len(nArgs) != 1 {
		log.Print("Usage: parsepb [-help] [...flags] /path/to/pathbuilder")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pb, err := pbxml.Load(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to load Pathbuilder: %s", err)
	}

	switch {
	case flagAscii: // format as text
		fmt.Println(pbtxt.Marshal(pb))
	case flagPretty: // format as pretty xml
		bytes, err := xml.MarshalIndent(pbxml.New(pb), "", "    ")
		if err != nil {
			log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
		}
		fmt.Println(string(bytes))
	default: // format as unpretty xml
		bytes, err := xml.Marshal(pbxml.New(pb))
		if err != nil {
			log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
		}
		fmt.Println(string(bytes))
	}
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
