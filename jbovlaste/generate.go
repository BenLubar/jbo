// +build ignore

package main

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/BenLubar/jbo/jbovlaste"
)

func main() {
	matches, err := filepath.Glob("*.xml")
	if err != nil {
		log.Panicln(err)
	}
	sort.Strings(matches)

	var dict jbovlaste.Dictionary
	for _, name := range matches {
		d, err := read(name)
		if err != nil {
			log.Panicln(name, err)
		}
		dict.Direction = append(dict.Direction, d.Direction...)
	}

	f, err := os.Create("jbovlaste.gen.go")
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()

	_, err = f.WriteString(`package jbovlaste

import (
	"encoding/xml"
	"strings"
	"sync"
)

var jbovlaste Dictionary
var once sync.Once

func all() *Dictionary {
	once.Do(func() {
		err := xml.NewDecoder(strings.NewReader(` + "`" + xml.Header)
	if err != nil {
		log.Panicln(err)
	}

	type dictionary jbovlaste.Dictionary

	enc := xml.NewEncoder(backtickWriter{f})
	enc.Indent("", "\t")
	err = enc.Encode((*dictionary)(&dict))
	if err != nil {
		log.Panicln(err)
	}

	_, err = f.WriteString("`" + `)).Decode(&jbovlaste)
		if err != nil {
			panic(err)
		}
	})

	return &jbovlaste
}
`)
	if err != nil {
		log.Panicln(err)
	}
}

func read(name string) (d jbovlaste.Dictionary, err error) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer errHelper(f.Close, &err)

	err = xml.NewDecoder(f).Decode(&d)
	return
}

func errHelper(f func() error, err *error) {
	if e := f(); *err == nil {
		*err = e
	}
}

var backtick = []byte("`")
var backtickEnt = []byte("&#96;")

type backtickWriter struct{ io.Writer }

func (w backtickWriter) Write(p []byte) (n int, err error) {
	var nn int

	for i, b := range bytes.Split(p, backtick) {
		if i != 0 {
			nn, err = w.Writer.Write(backtickEnt)
			if nn != len(backtickEnt) && err == nil {
				err = io.ErrShortWrite
			}
			if err != nil {
				return
			}
			n += len(backtick)
		}

		nn, err = w.Writer.Write(b)
		if nn != len(b) && err == nil {
			err = io.ErrShortWrite
		}
		n += nn
		if err != nil {
			return
		}
	}

	return
}
