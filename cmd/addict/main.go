// Command addict provides a graphical interface to the makeodbc command.
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ncruces/zenity"
	"github.com/tkw1536/FAU-CDI/drincw"
	"github.com/tkw1536/FAU-CDI/drincw/odbc"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
	"github.com/tkw1536/FAU-CDI/drincw/sql"
	"muzzammil.xyz/jsonc"
)

func main() {
	pb, ok := loadPathbuilder()
	if !ok {
		return
	}
	builder, ok := loadSelectors()
	if !ok {
		return
	}

	odbcs := odbc.MakeServer(pb)
	if err := builder.Apply(&odbcs); err != nil {
		zenity.Error(fmt.Sprintf("Unable to apply builder: %s", err))
	}

	doOdbc, doSelectors, doCancel := whatToDo()
	switch {
	case doCancel:
		return
	case doOdbc:
		ok = saveODBC(odbcs)
	case doSelectors:
		ok = saveSelectors(builder)
	}

	if !ok {
		return
	}
	zenity.Info("Finished!")
}

const SOURCE_SELECTOR_TITLE = "Where do you want to load the pathbuilder from?"
const SOURCE_SELECTOR_FILE = "File"
const SOURCE_SELECTOR_REMOTE = "Remote URL"
const SOURCE_FILE = "Which Pathbuilder do you want to load?"
const SOURCE_URL = "Which URL do you want to load?"

func loadPathbuilder() (pb pathbuilder.Pathbuilder, ok bool) {
	sTyp, err := zenity.List(SOURCE_SELECTOR_TITLE, []string{SOURCE_SELECTOR_FILE, SOURCE_SELECTOR_REMOTE}, zenity.Title(SOURCE_SELECTOR_TITLE))
	if err == zenity.ErrCanceled || (err == nil && sTyp == "") {
		return pb, false
	}
	if err != nil {
		log.Fatal(err)
	}

	var source string
	if sTyp == SOURCE_SELECTOR_FILE {
		source, err = zenity.SelectFile(zenity.Title(SOURCE_FILE))
	} else {
		source, err = zenity.Entry(SOURCE_URL, zenity.Title(SOURCE_URL))
	}

	if err == zenity.ErrCanceled || (err == nil && sTyp == "") {
		return pb, false
	}
	if err != nil {
		log.Fatal(err)
	}

	pb, err = pbxml.Load(source)
	if err != nil {
		zenity.Error(fmt.Sprintf("Unable to load pathbuilder from %s: %s", source, err))
		return pb, false
	}
	return pb, true
}

const LOAD_SELECTORS = "Do you want to load a selectors file?"
const SELECTOR_LOAD_FILE = "Which selector file do you want to load?"

func loadSelectors() (builder sql.Builder, ob bool) {
	err := zenity.Question(LOAD_SELECTORS, zenity.OKLabel("Yes"), zenity.CancelLabel("No"))
	if err == zenity.ErrCanceled {
		return builder, true
	}
	if err != nil {
		log.Fatal(err)
	}

	source, err := zenity.SelectFile(zenity.Title(SELECTOR_LOAD_FILE))
	if err == zenity.ErrCanceled {
		return builder, false
	}
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := os.ReadFile(source)
	if err != nil {
		zenity.Error(fmt.Sprintf("Unable to load selectors from %s: %s", source, err))
		return builder, false
	}
	if err := jsonc.Unmarshal(bytes, &builder); err != nil {
		zenity.Error(fmt.Sprintf("Unable to load selectors from %s: %s", source, err))
		return builder, false
	}

	return builder, true
}

const DO_TITLE = "What do you want to do?"
const DO_ODBC = "Create ODBC File"
const DO_SELECTORS = "Create Selectors File"

func whatToDo() (odbc bool, selectors bool, cancel bool) {
	act, err := zenity.List(SOURCE_SELECTOR_TITLE, []string{DO_ODBC, DO_SELECTORS}, zenity.Title(DO_TITLE))
	if err == zenity.ErrCanceled {
		return false, false, true
	}
	if err != nil {
		log.Fatal(err)
	}
	switch act {
	case DO_ODBC:
		return true, false, false
	case DO_SELECTORS:
		return false, true, false
	}
	return false, false, true
}

const SAVE_ODBC = "Where to save odbc file?"

func saveODBC(odbc odbc.Server) bool {
	path, err := zenity.SelectFileSave(zenity.Filename("import.xml"), zenity.Title(SAVE_ODBC))
	if err == zenity.ErrCanceled {
		return false
	}
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := xml.MarshalIndent(odbc, "", "    ")
	if err != nil {
		zenity.Error(fmt.Sprintf("Unable to marshal odbc: %s", err))
		return false
	}

	if err := os.WriteFile(path, bytes, os.ModePerm); err != nil {
		zenity.Error(fmt.Sprintf("Unable to save odbc to %s: %s", path, err))
		return false
	}

	return true
}

const SAVE_SELECTORS = "Where to save selectors file?"

func saveSelectors(builder sql.Builder) bool {
	path, err := zenity.SelectFileSave(zenity.Filename("selectors.jsonc"), zenity.Title(SAVE_SELECTORS))
	if err == zenity.ErrCanceled {
		return false
	}
	if err != nil {
		log.Fatal(err)
	}

	var buffer bytes.Buffer

	bytes, err := json.MarshalIndent(&builder, "", "    ")
	if err != nil {
		zenity.Error(fmt.Sprintf("Unable to marshal builder: %s", err))
		return false
	}
	buffer.WriteString(sql.MARSHAL_COMMENT_PREFIX + "\n")
	buffer.WriteString(string(bytes) + "\n")

	if err := os.WriteFile(path, buffer.Bytes(), os.ModePerm); err != nil {
		zenity.Error(fmt.Sprintf("Unable to save selectors to %s: %s", path, err))
		return false
	}

	return true
}

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
}
