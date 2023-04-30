package list

import (
    "fmt"
    "gopkg.in/yaml.v3"
    "os"
    "time"
)

type List struct {
    Title             string
    UpdatedAt         time.Time `yaml:"updated_at"`
    ReadmeGeneratedAt time.Time `yaml:"readme_generated_at"`
    Items             []*Item
}

func NewEmpty(title string) *List {
    return &List{
        Title: title,
    }
}

func NewFromFile(filename string) (l *List, err error) {
    _, err = os.Stat(filename)
    if os.IsNotExist(err) {
        return nil, err
    }
    bt, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    cfg := List{}
    err = yaml.Unmarshal(bt, &cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &cfg, nil
}

func (l *List) Save(filename string) error {
    bt, err := yaml.Marshal(l)
    if err != nil {
        return fmt.Errorf("failed to marshal list: %w", err)
    }
    err = os.WriteFile(filename, bt, 0644)
    if err != nil {
        return fmt.Errorf("failed to save list: %w", err)
    }
    return nil
}

func (l *List) ItemExists(item *Item) bool {
    for _, i := range l.Items {
        if i.Name == item.Name {
            return true
        }
    }
    return false
}

func (l *List) Add(item *Item) {
    l.Items = append(l.Items, item)
}

type Item struct {
    Name                 string    `yaml:"name"`
    Link                 string    `yaml:"link"`
    Description          string    `yaml:"description"`
    Ignore               bool      `yaml:"ignore"`
    IgnoreReason         string    `yaml:"ignore_reason"`
    Category             string    `yaml:"category"`
    AICategory           string    `yaml:"ai_category"`
    AICategoryConfidence float32   `yaml:"ai_category_confidence"`
    AIDescription        string    `yaml:"ai_description"`
    CreatedAt            time.Time `yaml:"created_at"`
    IsNew                bool      `yaml:"is_new"`
}

func (i *Item) String() string {
    return fmt.Sprintf("%s [%s(%f)] ignore=`%s`", i.Name, i.AICategory, i.AICategoryConfidence, i.IgnoreReason)
}
