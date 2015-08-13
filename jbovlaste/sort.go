// +build ignore

package main

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BenLubar/jbo/jbovlaste"
)

func main() {
	matches, err := filepath.Glob("*.xml")
	if err != nil {
		panic(err)
	}
	for _, m := range matches {
		var dict jbovlaste.Dictionary
		func() {
			f, err := os.Open(m)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if err := xml.NewDecoder(f).Decode(&dict); err != nil {
				panic(err)
			}
		}()

		for i := range dict.Direction {
			dir := &dict.Direction[i]
			for j := range dir.Valsi {
				valsi := &dir.Valsi[j]
				sort.Sort(glossSort(valsi.Gloss))
				sort.Sort(keywordSort(valsi.Keyword))
			}
			sort.Sort(natlangSort(dir.Natlang))
		}

		func() {
			f, err := os.Create(m)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if err := writeDictionary(f, &dict); err != nil {
				panic(err)
			}
		}()
	}
}

type glossSort []jbovlaste.Gloss
type keywordSort []jbovlaste.Keyword
type natlangSort []jbovlaste.Natlang

func (s glossSort) Len() int   { return len(s) }
func (s keywordSort) Len() int { return len(s) }
func (s natlangSort) Len() int { return len(s) }

func (s glossSort) Swap(i, j int)   { s[i], s[j] = s[j], s[i] }
func (s keywordSort) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s natlangSort) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s glossSort) Less(i, j int) bool {
	if s[i].Word < s[j].Word {
		return true
	}
	if s[i].Word > s[j].Word {
		return false
	}
	return s[i].Sense < s[j].Sense
}
func (s keywordSort) Less(i, j int) bool {
	if s[i].Place < s[j].Place {
		return true
	}
	if s[i].Place > s[j].Place {
		return false
	}
	if s[i].Word < s[j].Word {
		return true
	}
	if s[i].Word > s[j].Word {
		return false
	}
	return s[i].Sense < s[j].Sense
}
func (s natlangSort) Less(i, j int) bool {
	if s[i].Valsi < s[j].Valsi {
		return true
	}
	if s[i].Valsi > s[j].Valsi {
		return false
	}
	if s[i].Word < s[j].Word {
		return true
	}
	if s[i].Word > s[j].Word {
		return false
	}
	if s[i].Sense < s[j].Sense {
		return true
	}
	if s[i].Sense > s[j].Sense {
		return false
	}
	return s[i].Place < s[j].Place
}

func writeText(w io.Writer, s string) error {
	s = strings.Replace(s, "&", "&amp;", -1)
	s = strings.Replace(s, "\"", "&quot;", -1)
	s = strings.Replace(s, "'", "&apos;", -1)
	s = strings.Replace(s, "<", "&lt;", -1)
	s = strings.Replace(s, ">", "&gt;", -1)
	_, err := io.WriteString(w, s)
	return err
}

func writeDictionary(w io.Writer, dict *jbovlaste.Dictionary) error {
	if _, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="jbovlaste.xsl"?>
<dictionary>
`); err != nil {
		return err
	}

	for i := range dict.Direction {
		if err := writeDirection(w, &dict.Direction[i]); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, `</dictionary>





`); err != nil {
		return err
	}

	return nil
}

func writeDirection(w io.Writer, dir *jbovlaste.Direction) error {
	if _, err := io.WriteString(w, `<direction from="`); err != nil {
		return err
	}
	if err := writeText(w, dir.From); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `" to="`); err != nil {
		return err
	}
	if err := writeText(w, dir.To); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `">`); err != nil {
		return err
	}

	for i := range dir.Valsi {
		if err := writeValsi(w, &dir.Valsi[i]); err != nil {
			return err
		}
	}
	for i := range dir.Natlang {
		if err := writeNatlang(w, &dir.Natlang[i]); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, `</direction>`); err != nil {
		return err
	}

	return nil
}

func writeValsi(w io.Writer, valsi *jbovlaste.Valsi) error {
	if _, err := io.WriteString(w, `<valsi`); err != nil {
		return err
	}
	if valsi.Unofficial {
		if _, err := io.WriteString(w, ` unofficial="true"`); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, ` word="`); err != nil {
		return err
	}
	if _, err := io.WriteString(w, valsi.Word); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `" type="`); err != nil {
		return err
	}
	if _, err := io.WriteString(w, valsi.Type); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `">`); err != nil {
		return err
	}
	for _, s := range valsi.Rafsi {
		if _, err := io.WriteString(w, `
  <rafsi>`); err != nil {
			return err
		}
		if err := writeText(w, s); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `</rafsi>`); err != nil {
			return err
		}
	}
	if valsi.Selmaho != "" {
		if _, err := io.WriteString(w, `
  <selmaho>`); err != nil {
			return err
		}
		if err := writeText(w, valsi.Selmaho); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `</selmaho>`); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `
  <user>
    <username>`); err != nil {
		return err
	}
	if err := writeText(w, valsi.User.Name); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `</username>`); err != nil {
		return err
	}
	if valsi.User.Real != "" {
		if _, err := io.WriteString(w, `
    <realname>`); err != nil {
			return err
		}
		if err := writeText(w, valsi.User.Real); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `</realname>`); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `
  </user>`); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `
  <definition>`); err != nil {
		return err
	}
	if err := writeText(w, valsi.Definition); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `</definition>
  <definitionid>`); err != nil {
		return err
	}
	if _, err := io.WriteString(w, strconv.Itoa(valsi.DefinitionID)); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `</definitionid>`); err != nil {
		return err
	}
	if valsi.Notes != "" {
		if _, err := io.WriteString(w, `
  <notes>`); err != nil {
			return err
		}
		if err := writeText(w, valsi.Notes); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `</notes>`); err != nil {
			return err
		}
	}

	for i := range valsi.Gloss {
		if err := writeGloss(w, &valsi.Gloss[i]); err != nil {
			return nil
		}
	}
	for i := range valsi.Keyword {
		if err := writeKeyword(w, &valsi.Keyword[i]); err != nil {
			return nil
		}
	}

	if _, err := io.WriteString(w, `
</valsi>
`); err != nil {
		return err
	}

	return nil
}

func writeGloss(w io.Writer, gloss *jbovlaste.Gloss) error {
	if _, err := io.WriteString(w, `
  <glossword word="`); err != nil {
		return err
	}
	if err := writeText(w, gloss.Word); err != nil {
		return err
	}
	if gloss.Sense != "" {
		if _, err := io.WriteString(w, `" sense="`); err != nil {
			return err
		}
		if err := writeText(w, gloss.Sense); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `" />`); err != nil {
		return err
	}
	return nil
}

func writeKeyword(w io.Writer, keyword *jbovlaste.Keyword) error {
	if _, err := io.WriteString(w, `
  <keyword word="`); err != nil {
		return err
	}
	if err := writeText(w, keyword.Word); err != nil {
		return err
	}
	if _, err := io.WriteString(w, `" place="`); err != nil {
		return err
	}
	if _, err := io.WriteString(w, strconv.Itoa(keyword.Place)); err != nil {
		return err
	}
	if keyword.Sense != "" {
		if _, err := io.WriteString(w, `" sense="`); err != nil {
			return err
		}
		if err := writeText(w, keyword.Sense); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `" />`); err != nil {
		return err
	}
	return nil
}

func writeNatlang(w io.Writer, natlang *jbovlaste.Natlang) error {
	if _, err := io.WriteString(w, `<nlword word="`); err != nil {
		return err
	}
	if err := writeText(w, natlang.Word); err != nil {
		return err
	}
	if natlang.Sense != "" {
		if _, err := io.WriteString(w, `" sense="`); err != nil {
			return err
		}
		if err := writeText(w, natlang.Sense); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `" valsi="`); err != nil {
		return err
	}
	if err := writeText(w, natlang.Valsi); err != nil {
		return err
	}
	if natlang.Place != 0 {
		if _, err := io.WriteString(w, `" place="`); err != nil {
			return err
		}
		if _, err := io.WriteString(w, strconv.Itoa(natlang.Place)); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, `" />
`); err != nil {
		return err
	}
	return nil
}
