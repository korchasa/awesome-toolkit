package config

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "os"
    "strings"
)

type Config struct {
    Title        string
    Query        string
    DataPath     string `yaml:"data_path"`
    TempDataPath string `yaml:"temp_data_path"`
    ReadmePath   string `yaml:"readme_path"`
    Root         *CategoryDescription
}

func NewFromFile(filename string) (*Config, error) {
    bt, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    cfg := Config{}
    err = yaml.Unmarshal(bt, &cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &cfg, nil
}

type CategoryDescription struct {
    Title      string
    Prompt     string
    Categories []*CategoryDescription
}

func (d *CategoryDescription) TitlesTree(depth int) (t []string) {
    if d == nil {
        return nil
    }
    if d.Title != "" {
        t = append(t, strings.Repeat("    ", depth)+d.Title)
    }
    if len(d.Categories) > 0 {
        for _, sc := range d.Categories {
            t = append(t, sc.TitlesTree(depth+1)...)
        }
    }
    return t
}

func (d *CategoryDescription) FindTreeForm(title string) string {
    tree := d.TitlesTree(0)
    for _, t := range tree {
        if strings.Trim(t, " ") == strings.Trim(title, " ") {
            return t
        }
    }
    return ""
}

func (d *CategoryDescription) FindByTree(treeTitle string) string {
    return strings.Trim(treeTitle, " ")
}

func (d *CategoryDescription) Prompts() (p []string) {
    if d == nil {
        return nil
    }
    if d.Prompt != "" {
        p = append(p, d.Prompt)
    }
    if len(d.Categories) > 0 {
        for _, sc := range d.Categories {
            p = append(p, sc.Prompts()...)
        }
    }
    return p
}

func (d *CategoryDescription) FindTitleByPrompt(prompt string) string {
    if d == nil {
        return ""
    }
    if d.Title != "" {
        if strings.Trim(d.Prompt, " ") == strings.Trim(prompt, " ") {
            return d.Title
        }
    }
    if len(d.Categories) > 0 {
        for _, sc := range d.Categories {
            title := sc.FindTitleByPrompt(prompt)
            if title != "" {
                return title
            }
        }
    }
    return ""
}
