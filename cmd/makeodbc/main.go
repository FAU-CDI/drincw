// Command parsepb parses the pathbuilder and prints it again.
package main

// cSpell:words parsepb pathbuilder

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/FAU-CDI/drincw"
	"github.com/FAU-CDI/drincw/internal/sql"
	"github.com/FAU-CDI/drincw/odbc"
	"github.com/FAU-CDI/drincw/pathbuilder"
	"github.com/FAU-CDI/drincw/pathbuilder/pbxml"
	"muzzammil.xyz/jsonc"
)

func main() {
	if len(nArgs) != 1 {
		log.Print("Usage: makeodbc [-help] [...flags] /path/to/pathbuilder")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pb, err := pbxml.Load(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to load Pathbuilder: %s", err)
	}

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

	odbcs := odbc.NewServer(pb)
	if err := builder.Apply(&odbcs); err != nil {
		log.Fatalf("Unable to apply builder: %s", err)
	}

	switch {
	case flagDumpSQL != "":
		writeSQL(flagDumpSQL, pb, odbcs)
	case flagDumpSelectors:
		writeSelectors(builder)
	default:
		writeXML(odbcs)
	}
}

func writeSelectors(builder sql.Builder) {
	bytes, err := json.MarshalIndent(&builder, "", "    ")
	if err != nil {
		log.Fatalf("Unable to Marshal Builder: %s", err)
	}
	fmt.Println(string(sql.MARSHAL_COMMENT_PREFIX))
	fmt.Println(string(bytes))
}

func writeXML(odbc odbc.Server) {
	bytes, err := xml.MarshalIndent(odbc, "", "    ")
	if err != nil {
		log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
	}
	fmt.Println(string(bytes))
}

func writeSQL(name string, pb pathbuilder.Pathbuilder, odbc odbc.Server) {
	bundle := pb.Get(name)
	if bundle == nil {
		log.Fatalf("no such bundle: %s", name)
	}
	table := odbc.TableByID(bundle.Path.MachineName())
	if table.Name == "" {
		log.Fatalf("no table for: %s", flagDumpSQL)
	}
	fmt.Println(sql.ForTable(table))
}

var nArgs []string

var flagLoadSelectors string
var flagDumpSelectors bool
var flagDumpSQL string

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.StringVar(&flagLoadSelectors, "load-selectors", flagLoadSelectors, "load selector file")
	flag.BoolVar(&flagDumpSelectors, "dump-selectors", flagDumpSelectors, "generate a selectors template to generate sql statements")
	flag.StringVar(&flagDumpSQL, "sql", flagDumpSQL, "generate sql that the importer would run for the given bundle name")

	flag.Parse()
	nArgs = flag.Args()
}
