package config

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "os"
    "strings"
)

const (
    ConfigFilename   = "config.yaml"
    DataFilename     = ".data.yaml"
    TempDataFilename = ".data.tmp.yaml"
    ReadmeFilename   = "README.md"
)

type Config struct {
    Title   string
    Query   string
    Root    *CategoryDescription
    workDir string
}

func NewFromDir(dir string) (*Config, error) {
    bt, err := os.ReadFile(dir + "/" + ConfigFilename)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    cfg := Config{}
    err = yaml.Unmarshal(bt, &cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    cfg.workDir = dir
    return &cfg, nil
}

func (c *Config) Save() error {
    bt, err := yaml.Marshal(c)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    err = os.WriteFile(c.workDir+"/"+ConfigFilename, bt, 0644)
    if err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }
    return nil
}

func (c *Config) DataPath() string {
    return c.workDir + "/" + DataFilename
}

func (c *Config) TempDataPath() string {
    return c.workDir + "/" + TempDataFilename
}

func (c *Config) ReadmePath() string {
    return c.workDir + "/" + ReadmeFilename
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
