// +build ignore

package main

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/gob"
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

	var dictionary jbovlaste.Dictionary
	for _, name := range matches {
		d, err := read(name)
		if err != nil {
			log.Panicln(name, err)
		}
		dictionary.Direction = append(dictionary.Direction, d.Direction...)
	}

	f, err := os.Create("jbovlaste.gen.go")
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()

	_, err = f.WriteString(`package jbovlaste

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/gob"
	"strings"
	"sync"
)

var jbovlaste Dictionary
var once sync.Once

func all() *Dictionary {
	once.Do(func() {
		r, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(` + "`\n")
	if err != nil {
		log.Panicln(err)
	}

	err = compress(f, dictionary)
	if err != nil {
		log.Panicln(err)
	}

	_, err = f.WriteString("`" + `)))
		if err != nil {
			panic(err)
		}
		defer r.Close()

		err = gob.NewDecoder(r).Decode(&jbovlaste)
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

func compress(w io.Writer, d jbovlaste.Dictionary) (err error) {
	wrap := &wrapWriter{w: w, n: 79}
	defer errHelper(wrap.Close, &err)

	b64 := base64.NewEncoder(base64.StdEncoding, wrap)
	defer errHelper(b64.Close, &err)

	gz, err := gzip.NewWriterLevel(b64, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer errHelper(gz.Close, &err)

	err = gob.NewEncoder(gz).Encode(&d)
	return
}

func errHelper(f func() error, err *error) {
	if e := f(); *err == nil {
		*err = e
	}
}

var newline = []byte{'\n'}

// wrap every n bytes with a newline.
type wrapWriter struct {
	w io.Writer
	n int
	i int
}

func (w *wrapWriter) Write(p []byte) (n int, err error) {
	var nn int
	for w.i+len(p) > w.n {
		nn, err = w.w.Write(p[:w.n-w.i])
		n += nn
		if nn != w.n-w.i && err == nil {
			err = io.ErrShortWrite
		}
		if err != nil {
			return
		}
		nn, err = w.w.Write(newline)
		if nn != len(newline) && err == nil {
			err = io.ErrShortWrite
		}
		if err != nil {
			return
		}
		p = p[w.n-w.i:]
		w.i = 0
	}

	nn, err = w.w.Write(p)
	n += nn
	w.i += len(p)
	if nn != len(p) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

func (w *wrapWriter) Close() error {
	if w.i != 0 {
		n, err := w.w.Write(newline)
		if n != len(newline) && err == nil {
			err = io.ErrShortWrite
		}
		if err != nil {
			return err
		}
		w.i = 0
	}
	return nil
}
