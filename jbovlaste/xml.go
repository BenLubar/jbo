package jbovlaste

import (
	"sort"
	"sync"
)

type Dictionary struct {
	Direction []Direction `xml:"direction"`

	langs      []string
	langsIndex map[string]int
	langsOnce  sync.Once
}

func (d *Dictionary) initLangs() {
	d.langs = make([]string, 0, len(d.Direction)/2)
	d.langsIndex = make(map[string]int)

	for i := 0; i < len(d.Direction); i += 2 {
		lang := d.Direction[i].To
		d.langs = append(d.langs, lang)
		d.langsIndex[lang] = i
	}
}

func (d *Dictionary) Languages() []string {
	d.langsOnce.Do(d.initLangs)

	return d.langs
}

func (d *Dictionary) Language(name string) *Language {
	d.langsOnce.Do(d.initLangs)

	if i, ok := d.langsIndex[name]; ok {
		return &Language{
			Valsi:   d.Direction[i].Valsi,
			Natlang: d.Direction[i+1].Natlang,
		}
	}
	return nil
}

type Language struct {
	Valsi   []Valsi   `xml:"valsi"`
	Natlang []Natlang `xml:"nlword"`

	valsiTypes  []string
	valsiByType map[string][]string
	valsiIndex  map[string]int
	valsiOnce   sync.Once
}

func (l *Language) initValsi() {
	l.valsiByType = make(map[string][]string)
	l.valsiIndex = make(map[string]int)

	for i, v := range l.Valsi {
		l.valsiIndex[v.Word] = i
		l.valsiByType[v.Type] = append(l.valsiByType[v.Type], v.Word)
	}

	for t := range l.valsiByType {
		l.valsiTypes = append(l.valsiTypes, t)
	}
	sort.Strings(l.valsiTypes)
}

func (l *Language) WordTypes() []string {
	l.valsiOnce.Do(l.initValsi)

	return l.valsiTypes
}

func (l *Language) WordsByType(typ string) []string {
	l.valsiOnce.Do(l.initValsi)

	return l.valsiByType[typ]
}

func (l *Language) Word(lojban string) *Valsi {
	l.valsiOnce.Do(l.initValsi)

	if i, ok := l.valsiIndex[lojban]; ok {
		return &l.Valsi[i]
	}

	return nil
}

type Direction struct {
	From string `xml:"from,attr"`
	To   string `xml:"to,attr"`
	Language
}

type Valsi struct {
	Word         string    `xml:"word,attr"`
	Type         string    `xml:"type,attr"`
	Unofficial   bool      `xml:"unofficial,attr,omitempty"`
	Rafsi        []string  `xml:"rafsi"`
	Selmaho      string    `xml:"selmaho,omitempty"`
	User         User      `xml:"user"`
	Definition   string    `xml:"definition"`
	DefinitionID int       `xml:"definitionid"`
	Notes        string    `xml:"notes,omitempty"`
	Gloss        []Gloss   `xml:"glossword"`
	Keyword      []Keyword `xml:"keyword"`
}

type Natlang struct {
	Word  string `xml:"word,attr"`
	Sense string `xml:"sense,attr,omitempty"`
	Valsi string `xml:"valsi,attr"`
	Place int    `xml:"place,attr,omitempty"`
}

type User struct {
	Name string `xml:"username"`
	Real string `xml:"realname,omitempty"`
}

type Gloss struct {
	Word  string `xml:"word,attr"`
	Sense string `xml:"sense,attr,omitempty"`
}

type Keyword struct {
	Gloss
	Place int `xml:"place,attr"`
}
