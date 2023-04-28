package awesome_list

type List struct {
    Title      string
    Query      string
    Categories []*Category
}

func (l *List) ListCategoriesPrompts() (p []string) {
    for _, c := range l.Categories {
        p = append(p, c.Prompts()...)
    }
    return p
}

func (l *List) AddItemByTitlePrompt(prompt string, item Item) (found bool) {
    for _, c := range l.Categories {
        if c.AddItemByPrompt(prompt, item) {
            return true
        }
    }
    return false

}

func (l *List) ItemExists(uri string) bool {
    for _, c := range l.Categories {
        if c.ItemExists(uri) {
            return true
        }
    }
    return false
}

type Category struct {
    Title      string
    Prompt     string      `yaml:"prompt,omitempty"`
    Categories []*Category `yaml:"subcategories,omitempty"`
    Items      []*Item
}

func (c *Category) Prompts() (p []string) {
    if c.Prompt != "" {
        p = append(p, c.Prompt)
    }
    if len(c.Categories) > 0 {
        for _, sc := range c.Categories {
            p = append(p, sc.Prompts()...)
        }
    }
    return p
}

func (c *Category) AddItemByPrompt(prompt string, item Item) bool {
    if c.Prompt == prompt {
        c.Items = append(c.Items, &item)
        return true
    }
    if len(c.Categories) > 0 {
        for _, sc := range c.Categories {
            if sc.AddItemByPrompt(prompt, item) {
                return true
            }
        }
    }
    return false
}

type Item struct {
    Name          string
    Link          string
    Description   string
    AIDescription string
}
