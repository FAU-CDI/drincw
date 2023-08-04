// Command pbfmt formats a pathbuilder and prints it again
package main

// cSpell:words pbfmt pathbuilder

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/FAU-CDI/drincw"
	"github.com/FAU-CDI/drincw/pathbuilder/pbtxt"
	"github.com/FAU-CDI/drincw/pathbuilder/pbxml"
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
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.BoolVar(&flagAscii, "ascii", flagAscii, "format as text instead of xml")
	flag.BoolVar(&flagPretty, "pretty", flagPretty, "format as prettified xml")

	flag.Parse()
	nArgs = flag.Args()
}
