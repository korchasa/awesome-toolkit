package adder

import (
    "context"
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/AlecAivazis/survey/v2/terminal"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/repo_classifier"
    log "github.com/sirupsen/logrus"
    "os"
    "time"
)

type App struct {
    classifier *repo_classifier.RepoClassifier
    cfg        *config.Config
    data       *list.List
}

func MustBuildApp(aiToken string, cfg *config.Config) *App {
    return &App{
        classifier: repo_classifier.NewRepoClassifier(aiToken, cfg),
        cfg:        cfg,
        data:       mustLoadData(cfg),
    }
}

func mustLoadData(cfg *config.Config) *list.List {
    data, err := list.NewFromFile(cfg.DataPath())
    if os.IsNotExist(err) {
        log.Infof("data file `%s` not found, creating new one", cfg.DataPath())
        data = list.NewEmpty(cfg.Title)
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
        err = s.data.Save(s.cfg.DataPath())
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
            Name:     "Description",
            Prompt:   &survey.Multiline{Message: "Enter the description:"},
            Validate: survey.Required,
        },
        {
            Name: "Category",
            Prompt: &survey.Select{
                Message:  "Enter the category:",
                Options:  s.cfg.Root.TitlesTree(0),
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
