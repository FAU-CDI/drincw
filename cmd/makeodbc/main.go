// Command parsepb parses the pathbuilder and prints it again.
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/sql"
	"muzzammil.xyz/jsonc"
)

func main() {
	if len(nArgs) != 1 {
		log.Print("Usage: makeodbc [-help] [...flags] /path/to/pathbuilder")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pbx, err := drincw.LoadPathbuilderXML(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to load Pathbuilder: %s", err)
	}
	pb := pbx.Pathbuilder()
	odbc := pb.ODBC()

	var builder sql.Builder
	if flagLoadSelectors != "" {
		bytes, err := os.ReadFile(flagLoadSelectors)
		if err != nil {
			log.Fatalf("Unable to load Selectors: %s", err)
		}
		if err := jsonc.Unmarshal(bytes, &builder); err != nil {
			log.Fatalf("Unable to load Selectors: %s", err)
		}
	} else {
		builder = sql.NewBuilder(pb)
	}

	if err := builder.Apply(&odbc); err != nil {
		log.Fatalf("Unable to apply builder: %s", err)
	}

	switch {
	case flagDumpSelectors:
		writeSelectors(builder)
	default:
		writeXML(odbc)
	}
}

func writeSelectors(builder sql.Builder) {
	bytes, err := json.MarshalIndent(&builder, "", "    ")
	if err != nil {
		log.Fatalf("Unable to Marshal Builder: %s", err)
	}
	fmt.Println(string(bytes))
}

func writeXML(odbc drincw.ODBCServer) {
	bytes, err := xml.MarshalIndent(odbc, "", "    ")
	if err != nil {
		log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
	}
	fmt.Println(string(bytes))
}

var nArgs []string

var flagLoadSelectors string
var flagDumpSelectors bool

func init() {
	flag.StringVar(&flagLoadSelectors, "load-selectors", flagLoadSelectors, "load selector file")
	flag.BoolVar(&flagDumpSelectors, "dump-selectors", flagDumpSelectors, "generate a selectors template to generate sql statements")
	flag.Parse()
	nArgs = flag.Args()
}
