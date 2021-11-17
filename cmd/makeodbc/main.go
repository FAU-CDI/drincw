// Command parsepb parses the pathbuilder and prints it again.
package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tkw1536/FAU-CDI/drincw"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: makeodbc /path/to/pathbuilder")
	}

	content, err := open(os.Args[1])
	if err != nil {
		log.Fatalf("Unable to open pathbuilder: %s", err)
	}

	var pb drincw.XMLPathbuilder
	if err := xml.Unmarshal(content, &pb); err != nil {
		log.Fatalf("Unable to Unmarshal Pathbuilder: %s", err)
	}

	bytes, err := xml.MarshalIndent(pb.Pathbuilder().ODBC(), "", "    ")
	if err != nil {
		log.Fatalf("Unable to Marshal Pathbuilder: %s", err)
	}
	fmt.Println(string(bytes))
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
