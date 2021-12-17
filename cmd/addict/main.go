package main

import (
	"fmt"
	"log"

	"github.com/ncruces/zenity"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder"
	"github.com/tkw1536/FAU-CDI/drincw/pathbuilder/pbxml"
)

func main() {
	pb, ok := loadPathbuilder()
	fmt.Printf("%#v %#v", pb, ok)
}

const SOURCE_SELECTOR_TITLE = "Where do you want to load the pathbuilder from?"
const SOURCE_SELECTOR_FILE = "File"
const SOURCE_SELECTOR_REMOTE = "Remote URL"
const SOURCE_FILE = "Which Pathbuilder do you want to load?"
const SOURCE_URL = "Which URL do you want to load?"

func loadPathbuilder() (pb pathbuilder.Pathbuilder, ok bool) {
	sTyp, err := zenity.List(SOURCE_SELECTOR_TITLE, []string{SOURCE_SELECTOR_FILE, SOURCE_SELECTOR_REMOTE}, zenity.Title(SOURCE_SELECTOR_TITLE), zenity.DisallowEmpty())
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
