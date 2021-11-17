package main

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/tkw1536/FAU-CDI/drincw"
)

func main() {
	if len(os.Args) <= 1 {
		panic("not enough args")
	}

	bytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var pb drincw.PathbuilderInterface
	if err := xml.Unmarshal(bytes, &pb); err != nil {
		panic(err)
	}

	xmls, err := xml.MarshalIndent(pb.BundleDict(), "", "   ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(xmls))
}
