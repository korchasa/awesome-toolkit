package ignorer

import (
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "unicode"
)

const maxNonLatinSymbolsPercent = 0.1

type Ignorer struct {
}

func NewIgnorer() *Ignorer {
    return &Ignorer{}
}

func (i *Ignorer) ResolveIgnores(item *list.Item, readme string) {
    if nonLatinPercentage(item.Description+readme) > maxNonLatinSymbolsPercent {
        item.Ignore = true
        item.IgnoreReason = "non latin description"
    }
    return
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
            if unicode.In(r, unicode.Latin) == false {
                nonLatinCount++
            }
        }
    }
    if totalChars == 0 {
        return 0
    }
    return float64(nonLatinCount) / float64(totalChars) * 100
}
