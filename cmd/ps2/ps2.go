// Command ps2 generates a sparql query for a specific field
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/FAU-CDI/drincw"
	"github.com/FAU-CDI/drincw/pathbuilder"
	"github.com/FAU-CDI/drincw/pathbuilder/pbxml"
)

func main() {
	if len(nArgs) != 2 {
		log.Print("Usage: ps2 [-help] [...flags] /path/to/pathbuilder field")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pb, err := pbxml.Load(nArgs[0])
	if err != nil {
		log.Fatalf("Unable to load Pathbuilder: %s", err)
	}

	var field *pathbuilder.Path
	for _, path := range pb.Paths() {
		if path.ID == nArgs[1] {
			field = &path
			break
		}
	}
	if field == nil {
		log.Fatalf("Unable to load field")
	}

	var builder strings.Builder
	var counter int

	var last, current string
	current = "?v0"
	for i, path := range field.PathArray {
		counter++
		if i%2 == 0 {
			fmt.Fprintf(&builder, "        %s a <%s> .\n", current, path)
			continue
		}

		last, current = current, fmt.Sprintf("?v%d", counter)
		fmt.Fprintf(&builder, "        %s <%s> %s .\n", last, path, current)
	}

	if field.DatatypeProperty != "" {
		last, current = current, fmt.Sprintf("?v%d", counter)
		fmt.Fprintf(&builder, "        %s <%s> %s .\n", last, field.DatatypeProperty, current)
	}

	fmt.Println("SELECT * WHERE {")
	fmt.Println("    GRAPH ?g {")
	fmt.Print(builder.String())
	fmt.Println("    }")
	fmt.Println("}")
}

var nArgs []string

func init() {
	var legalFlag bool = false
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Display legal notices and exit")
	defer func() {
		if legalFlag {
			fmt.Print(drincw.LegalText())
			os.Exit(0)
		}
	}()

	flag.Parse()
	nArgs = flag.Args()
}
