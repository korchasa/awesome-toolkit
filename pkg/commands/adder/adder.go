package adder

import (
    "context"
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/AlecAivazis/survey/v2/terminal"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/repo_classifier"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "os"
    "time"
)

type App struct {
    classifier   *repo_classifier.RepoClassifier
    categoryTree *config.CategoryDescription
    data         *list.List
    dataPath     string
}

func MustBuildApp(ai *openai.Client, cfg *config.Config) *App {
    return &App{
        classifier:   repo_classifier.NewRepoClassifier(ai, cfg.Root),
        categoryTree: cfg.Root,
        data:         mustLoadData(cfg.DataPath()),
        dataPath:     cfg.DataPath(),
    }
}

func mustLoadData(path string) *list.List {
    data, err := list.NewFromFile(path)
    if os.IsNotExist(err) {
        log.Infof("data file `%s` not found, creating new one", path)
        data = list.NewEmpty()
    } else if err != nil {
        log.Fatalf("failed to load data: %s", err)
    }
    return data
}

func (s *App) Run(_ context.Context) error {
    for {
        item, err := s.askForItem()
        if err != nil {
            return fmt.Errorf("failed to ask for item: %w", err)
        }

        s.data.Add(item)
        err = s.data.Save(s.dataPath)
        if err != nil {
            return fmt.Errorf("failed to save data: %w", err)
        }

        if !askForContinue() {
            break
        }
    }

    return nil
}

func (s *App) askForItem() (*list.Item, error) {
    item := list.Item{
        IsNew:     true,
        CreatedAt: time.Now(),
    }

    lp := []*survey.Question{
        {
            Name:     "Link",
            Prompt:   &survey.Input{Message: "Enter the link:"},
            Validate: survey.Required,
        },
    }
    err := survey.Ask(lp, &item)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    if s.data.ItemExists(&item) {
        log.Warnf("item `%s` already exists", item.Link)
        return s.askForItem()
    }

    prompts := []*survey.Question{
        {
            Name:     "Name",
            Prompt:   &survey.Input{Message: "Enter the name:"},
            Validate: survey.Required,
        },
        {
            Name:   "Description",
            Prompt: &survey.Multiline{Message: "Enter the description:"},
        },
        {
            Name: "Category",
            Prompt: &survey.Select{
                Message:  "Enter the category:",
                Options:  s.categoryTree.TitlesTree(0),
                PageSize: 30,
            },
            Validate: survey.Required,
        },
    }
    err = survey.Ask(prompts, &item)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    item.Category = s.categoryTree.FindByTree(item.Category)

    return &item, nil
}

func askForContinue() bool {
    var qs = []*survey.Question{
        {
            Name: "Confirm",
            Prompt: &survey.Confirm{
                Message: "Add another item?",
                Default: true,
            },
        },
    }
    answer := struct{ Confirm bool }{}
    err := survey.Ask(qs, &answer)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    return answer.Confirm
}
