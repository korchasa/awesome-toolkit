package readme_generator

import (
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "os"
    "sort"
    "strings"
)

const BodyPlaceholder = "{{body}}"

type ReadmeGenerator struct {
}

func NewReadmeGenerator() *ReadmeGenerator {
    return &ReadmeGenerator{}
}

func (r *ReadmeGenerator) Generate(cfg *config.Config, list *list.List) (content string, err error) {
    tpl, err := os.ReadFile(cfg.ReadmeTemplatePath())
    if err != nil {
        return "", fmt.Errorf("failed to read template `%s`: %w", cfg.ReadmeTemplatePath(), err)
    }
    content = string(tpl)
    if !strings.Contains(content, BodyPlaceholder) {
        return "", fmt.Errorf("template `%s` does not contain `%s`", cfg.ReadmeTemplatePath(), BodyPlaceholder)
    }

    sort.Slice(list.Items, func(i, j int) bool {
        return list.Items[i].Name < list.Items[j].Name
    })

    tocBody := ""
    itemsBody := ""
    for _, category := range cfg.Root.Categories {
        tocBody += genSublistTOC(category, list, 1)
        itemsBody += genSublist(category, list, 1)
    }

    content = strings.Replace(content, BodyPlaceholder, tocBody+itemsBody, 1)
    return content, nil
}

func genSublistTOC(category *config.CategoryDescription, l *list.List, i int) string {

    content := fmt.Sprintf(
        "%s- [%s](#%s)\n",
        strings.Repeat("  ", i),
        category.Title,
        category.Title)
    for _, sc := range category.Categories {
        content += genSublistTOC(sc, l, i+1)
    }
    return content
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
