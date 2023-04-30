package readme_generator

import (
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "strings"
)

type ReadmeGenerator struct {
}

func NewReadmeGenerator() *ReadmeGenerator {
    return &ReadmeGenerator{}
}

func (r *ReadmeGenerator) Generate(cfg *config.Config, list *list.List) (content string, err error) {
    content = fmt.Sprintf("# %s\n\n", list.Title)
    for _, category := range cfg.Root.Categories {
        content += genSublist(category, list, 1)
    }
    return content, nil
}

func genSublist(cat *config.CategoryDescription, list *list.List, intend int) string {
    content := fmt.Sprintf("%s %s\n\n", strings.Repeat("#", intend+1), cat.Title)
    for _, item := range list.Items {
        if item.Ignore || item.Category != cat.Title {
            continue
        }
        content += fmt.Sprintf("- [%s](%s) - %s\n", item.Name, item.Link, item.Description)
    }
    content += "\n"
    for _, sc := range cat.Categories {
        content += genSublist(sc, list, intend+1)
    }
    return content
}
