package ignorer

import (
	"github.com/korchasa/awesome-toolkit/pkg/list"
	log "github.com/sirupsen/logrus"
	"strings"
	"unicode"
)

const maxNonLatinSymbolsPercent = 3

type Ignorer struct {
}

func NewIgnorer() *Ignorer {
	return &Ignorer{}
}

func (i *Ignorer) ResolveIgnores(item *list.Item, readme string) {
	text := item.Description + readme
	skipSyms := []string{"`", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ".", "!", "|"}
	for _, sym := range skipSyms {
		text = strings.ReplaceAll(text, sym, " ")
	}
	p := nonLatinPercentage(text)
	if p > maxNonLatinSymbolsPercent {
		log.Infof("not-english `%s` (%f)", item.Link, p)
		item.Ignore = true
		item.IgnoreReason = "not-english"
	}
}

func nonLatinPercentage(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	nonLatinCount := 0
	totalChars := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			totalChars++
			if !unicode.In(r, unicode.Latin) {
				nonLatinCount++
			}
		}
	}
	if totalChars == 0 {
		return 0
	}
	return float64(nonLatinCount) / float64(totalChars) * 100
}
