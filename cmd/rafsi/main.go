package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/BenLubar/jbo/jbovlaste"
)

func main() {
	bangu := jbovlaste.All().Language("English")

	// place number of first place after conversion
	se := map[string]int{
		"se":    2,
		"te":    3,
		"ve":    4,
		"xe":    5,
		"to'ai": 3,
		"vo'ai": 4,
		"xo'ai": 5,
	}

	mathMatch := regexp.MustCompile(`\$(.+?)\$`)
	mathReplace := strings.NewReplacer("{", "", "}", "", "*", "×", "/", "÷").Replace
	supMatch := regexp.MustCompile(`\^([\d+-]+)`)
	supReplace := strings.NewReplacer("^", "", "0", "⁰", "1", "¹", "2", "²", "3", "³", "4", "⁴", "5", "⁵", "6", "⁶", "7", "⁷", "8", "⁸", "9", "⁹", "+", "⁺", "-", "⁻").Replace
	subMatch := regexp.MustCompile(`_([\d+-]+)`)
	subReplace := strings.NewReplacer("_", "", "0", "₀", "1", "₁", "2", "₂", "3", "₃", "4", "₄", "5", "₅", "6", "₆", "7", "₇", "8", "₈", "9", "₉", "+", "₊", "-", "₋").Replace

	word := func(s string, place int) {
		valsi := bangu.Word(s)

		if valsi == nil {
			fmt.Print(s, " (!!UNKNOWN WORD!!)")
			return
		}
		if place == 0 {
			place = 1
		}
		if place == -1 {
			if len(valsi.Gloss) != 0 {
				fmt.Print(s, " (", valsi.Gloss[0].Word)
				if valsi.Gloss[0].Sense != "" {
					fmt.Print(", ", valsi.Gloss[0].Sense)
				}
				fmt.Print(")")
				return
			}
		} else {
			for _, k := range valsi.Keyword {
				if k.Place != place {
					continue
				}
				fmt.Print(s, " (", k.Word)
				if k.Sense != "" {
					fmt.Print(", ", k.Sense)
				}
				fmt.Print(")")
				return
			}
		}

		fmt.Print(s, " (!!NO GLOSSARY ENTRY!!", mathMatch.ReplaceAllStringFunc(valsi.Definition, func(m string) string {
			m = m[1 : len(m)-1]
			m = mathReplace(m)
			m = supMatch.ReplaceAllStringFunc(m, supReplace)
			m = subMatch.ReplaceAllStringFunc(m, subReplace)
			return m
		}), "!!)")
	}

	for _, lujvo := range bangu.WordsByType("lujvo") {
		rafsi := splitRafsi(lujvo)

		place := 0

		word(lujvo, -1)
		fmt.Print(" = ")
		for i, r := range rafsi {
			if i != 0 {
				fmt.Print(" + ")
			}
			word(r, place)
			place = se[r] // XXX FIXME: this doesn't work with multiple cmavo in a row.
		}
		fmt.Println()
		fmt.Println()
	}
}

var rafsi = make(map[string]string)
var rafsiOnce sync.Once

func splitRafsi(compound string) []string {
	// shamelessly stolen from github.com/dag/jbo (python command)
	const (
		c       = `[bcdfgjklmnprstvxz]`
		v       = `[aeiou]`
		cc      = `(?:bl|br|cf|ck|cl|cm|cn|cp|cr|ct|dj|dr|dz|fl|fr|gl|gr|jb|jd|jg|jm|jv|kl|kr|ml|mr|pl|pr|sf|sk|sl|sm|sn|sp|sr|st|tc|tr|ts|vl|vr|xl|xr|zb|zd|zg|zm|zv)`
		vv      = `(?:ai|ei|oi|au)`
		rafsi3v = `(?:` + cc + v + `|` + c + vv + `|` + c + v + `'` + v + `)`
		rafsi3  = `(?:` + rafsi3v + `|` + c + v + c + `)`
		rafsi4  = `(?:` + c + v + c + c + `|` + cc + v + c + `)`
		rafsi5  = rafsi4 + v
	)

	rafsiOnce.Do(func() {
		bangu := jbovlaste.All().Language("English")

		for i := range bangu.Valsi {
			valsi := &bangu.Valsi[i]

			for _, r := range valsi.Rafsi {
				rafsi[r] = valsi.Word
			}
		}

		for _, s := range bangu.WordsByType("gismu") {
			rafsi[s[:4]] = s
		}

		// unofficial stuff:
		for _, s := range bangu.WordsByType("experimental gismu") {
			if s == "datru" {
				// hard-coded to ignore outdated form of datro.
				continue
			}

			if r, ok := rafsi[s[:4]]; ok {
				panic("duplicate rafsi " + s[:4] + " (" + s + ", " + r + ")")
			}
			rafsi[s[:4]] = s
		}

		re := regexp.MustCompile(`\W-(` + rafsi3 + `)-(?:\W|\z)`)

		for i := range bangu.Valsi {
			valsi := &bangu.Valsi[i]

			if valsi.Type == "lujvo" || valsi.Type == "fu'ivla" || strings.HasPrefix(valsi.Type, "obsolete ") {
				continue
			}

			if valsi.Word == "zi'ai" {
				// hard-coded to ignore false positive
				continue
			}

			for _, m := range re.FindAllStringSubmatch(bangu.Valsi[i].Notes, -1) {
				if s, ok := rafsi[m[1]]; ok {
					panic("duplicate rafsi " + m[1] + " (" + valsi.Word + ", " + s + ")")
				}

				rafsi[m[1]] = valsi.Word
			}
		}
	})

	// shamelessly stolen from github.com/dag/jbo (python command)
	for i := 1; i <= len(compound)/3; i++ {
		var (
			reg  = strings.Repeat(`(?:(`+rafsi3+`)[nry]??|(`+rafsi4+`)y)`, i)
			reg2 = `\A` + reg + `(` + rafsi3v + `|` + rafsi5 + `)\z`
			re   = regexp.MustCompile(reg2)
		)
		matches := re.FindStringSubmatch(compound)
		if matches != nil {
			filtered := make([]string, 0, i+1)
			for _, s := range matches[1:] {
				if s != "" {
					if ss, ok := rafsi[s]; ok {
						s = ss
					}
					filtered = append(filtered, s)
				}
			}
			return filtered
		}
	}

	return nil
}
