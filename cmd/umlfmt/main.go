package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"fknsrs.biz/p/plantuml/parser"
)

var (
	list  bool
	write bool
)

func init() {
	flag.BoolVar(&list, "l", false, "list files whose formatting differs from umlfmt's")
	flag.BoolVar(&write, "w", false, "write result to (source) file instead of stdout")
}

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

		buf := bytes.NewBuffer(nil)
		if err := parser.FormatDocument(*doc, buf); err != nil {
			log.Printf("error formatting %s: %s\n", f, err)
			continue
		}
		res := buf.Bytes()

		if bytes.Equal(bytes.TrimSpace(src), bytes.TrimSpace(res)) {
			continue
		}

		if list {
			fmt.Println(f)
			continue
		}

		if write {
			if err := ioutil.WriteFile(f, res, 0644); err != nil {
				log.Printf("error writing %s: %s\n", f, err)
				continue
			}
		} else {
			io.Copy(os.Stdout, bytes.NewBuffer(res))
		}
	}
}
