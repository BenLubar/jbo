// +build ignore

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	const botKey = "z2BsnKYJhAB0VNsl" // https://github.com/lojban/jbovlaste/blob/master/export/xml-export.html#L25

	r, err := http.Get("http://jbovlaste.lojban.org/export/xml.html")
	if err != nil {
		log.Panicln(err)
	}
	if r.StatusCode != http.StatusOK {
		log.Panicln("xml.html request failed:", r.Status)
	}
	n, err := html.Parse(r.Body)
	if e := r.Body.Close(); err == nil {
		err = e
	}
	if err != nil {
		log.Panicln(err)
	}

	links := htmlFind(n, atom.Html, atom.Body, atom.Table, atom.Tbody, atom.Tr, atom.Td, atom.Ul, atom.Li, atom.A)
	if len(links) == 0 {
		log.Panicln("jbovlaste must have changed - scraping code failed")
	}
	for _, l := range links {
		const linkPrefix = "xml-export.html?lang="

		if l.FirstChild.Type != html.TextNode || len(l.Attr) != 1 || l.Attr[0].Namespace != "" || l.Attr[0].Key != "href" || !strings.HasPrefix(l.Attr[0].Val, linkPrefix) {
			log.Panicln("jbovlaste must have changed - link precondition failed")
		}

		func(code, name string) {
			const urlPrefix = "http://jbovlaste.lojban.org/export/xml-export.html?bot_key=" + botKey + "&lang="
			const fileSuffix = ".xml"

			log.Println("getting", name, "("+code+fileSuffix+")...")

			f, err := os.Create(code + fileSuffix)
			if err != nil {
				log.Panicln(err)
			}
			defer f.Close()

			r, err := http.Get(urlPrefix + code)
			if err != nil {
				log.Panicln(err)
			}
			defer r.Body.Close()
			if r.StatusCode != http.StatusOK {
				log.Panicln("request failed:", r.Status)
			}

			_, err = io.Copy(f, r.Body)
			if err != nil {
				log.Panicln(err)
			}
		}(l.Attr[0].Val[len(linkPrefix):], l.FirstChild.Data)
	}
}

func htmlFind(n *html.Node, atoms ...atom.Atom) []*html.Node {
	if len(atoms) == 0 {
		return []*html.Node{n}
	}
	var nodes []*html.Node
	for n = n.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.DataAtom == atoms[0] {
			nodes = append(nodes, htmlFind(n, atoms[1:]...)...)
		}
	}
	return nodes
}
