package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/FAU-CDI/drincw/internal/sql"
	"github.com/FAU-CDI/drincw/odbc"
)

func main() {
	if len(os.Args) != 3 {
		log.Print("Usage: dummysql /path/to/odbc tablename")
	}

	var server odbc.Server
	{
		odbcPath := os.Args[1]
		bytes, err := os.ReadFile(odbcPath)
		if err != nil {
			log.Fatalf("unable to open %q: %s", odbcPath, err)
		}
		if err := xml.Unmarshal(bytes, &server); err != nil {
			log.Fatalf("unable to open %q: %s", odbcPath, err)
		}
	}

	var sqls string
	{
		tableid := os.Args[2]

		var ok bool
		var table odbc.Table
		for _, t := range server.Tables {
			if t.Name == tableid {
				table = t
				ok = true
				break
			}
		}
		if !ok {
			log.Fatalf("no table %s", tableid)
		}
		sqls = sql.ForTable(table)
	}

	fmt.Println(sqls)
}
