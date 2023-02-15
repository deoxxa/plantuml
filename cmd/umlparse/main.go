package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"fknsrs.biz/p/plantuml/parser"
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stderr)

	for _, f := range flag.Args() {
		src, err := ioutil.ReadFile(f)
		if err != nil {
			log.Printf("error reading %s: %s\n", f, err)
			continue
		}

		doc, err := parser.ParseDocument(string(src))
		if err != nil {
			log.Printf("error parsing %s: %s\n", f, err)
			continue
		}

		spew.Dump(doc)
	}
}
