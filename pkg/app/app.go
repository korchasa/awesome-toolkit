package app

import (
    "context"
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/AlecAivazis/survey/v2/terminal"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/github"
    "github.com/korchasa/awesome-toolkit/pkg/ignorer"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/readme_generator"
    "github.com/korchasa/awesome-toolkit/pkg/repo_classifier"
    log "github.com/sirupsen/logrus"
    "os"
    "time"
)

type App struct {
    github     *github.GitHub
    classifier *repo_classifier.RepoClassifier
    cfg        *config.Config
    tempData   *list.List
    ignorer    *ignorer.Ignorer
}

func MustBuildApp(ghToken string, aiToken string, cfg *config.Config) *App {
    return &App{
        github:     github.NewGitHub(ghToken),
        classifier: repo_classifier.NewRepoClassifier(aiToken, cfg),
        cfg:        cfg,
        tempData:   mustLoadTempData(cfg),
        ignorer:    ignorer.NewIgnorer(),
    }
}

func mustLoadTempData(cfg *config.Config) *list.List {
    tempData, err := list.NewFromFile(cfg.TempDataPath())
    if os.IsNotExist(err) {
        log.Infof("temp data file `%s` not found, creating new one", cfg.TempDataPath())
        oldData, err := list.NewFromFile(cfg.DataPath())
        if os.IsNotExist(err) {
            log.Infof("data file `%s` not found, creating new one", cfg.DataPath())
            oldData = list.NewEmpty(cfg.Title)
        } else if err != nil {
            log.Fatalf("failed to load old data: %s", err)
        }
        tmp := *oldData
        tempData = &tmp
    } else if err != nil {
        log.Fatalf("failed to load temp data: %s", err)
    }
    return tempData
}

func (s *App) Run(ctx context.Context) error {
    items, err := s.github.SearchRepos(ctx, s.cfg.Query)
    if err != nil {
        return fmt.Errorf("failed to search repositories: %w", err)
    }
    count := len(items)
    for i, item := range items {
        stop, err := s.processFoundRepo(ctx, item, i, count)
        if err != nil {
            return fmt.Errorf("failed to process repo `%s`: %w", item.Name, err)
        }
        if stop {
            break
        }
        s.tempData.Add(item)
        if err := s.tempData.Save(s.cfg.TempDataPath()); err != nil {
            return fmt.Errorf("failed to save temp data: %w", err)
        }
    }
    if s.confirmDataReplacement() {
        log.Infof("saving data to `%s`", s.cfg.DataPath())
        s.tempData.UpdatedAt = time.Now()
        err = s.tempData.Save(s.cfg.DataPath())
        if err != nil {
            return fmt.Errorf("failed to save data: %w", err)
        }
        if err := os.Remove(s.cfg.TempDataPath()); err != nil {
            return fmt.Errorf("failed to remove temp data: %w", err)
        }
    }
    if s.confirmGenerateReadme() {
        oldData, err := list.NewFromFile(s.cfg.DataPath())
        if err != nil {
            log.Fatalf("failed to load old data: %s", err)
        }
        readme, err := readme_generator.NewReadmeGenerator().Generate(s.cfg, oldData)
        if err != nil {
            log.Fatalf("failed to generate readme: %s", err)
        }
        err = os.WriteFile(s.cfg.ReadmePath(), []byte(readme), 0644)
        if err != nil {
            log.Fatalf("failed to write readme: %s", err)
        }
        oldData.ReadmeGeneratedAt = time.Now()
        err = oldData.Save(s.cfg.DataPath())
        if err != nil {
            return fmt.Errorf("failed to save data: %w", err)
        }
    }
    return nil
}

func (s *App) processFoundRepo(ctx context.Context, item *list.Item, index int, count int) (stop bool, err error) {
    if s.tempData.ItemExists(item) {
        log.Infof("Skip `%s` because it already exists in data", item.Name)
        return false, nil
    }
    readme, err := s.github.GetReadme(ctx, item)
    if err != nil {
        return true, fmt.Errorf("failed to get readme for `%s`: %w", item.Name, err)
    }
    s.ignorer.ResolveIgnores(item, readme)
    if item.Ignore {
        log.Infof("Skip `%s` because `%s`", item.Name, item.IgnoreReason)
        return false, nil
    }
    err = s.classifier.ClassifyRepo(ctx, item, readme)
    if err != nil {
        return true, fmt.Errorf("failed to classify repo `%s`: %w", item.Name, err)
    }
    exit, err := s.askForCategory(item, index, count)
    if err != nil {
        return false, fmt.Errorf("failed to ask for category: %w", err)
    }
    if exit {
        return true, nil
    }
    return false, nil
}

func (s *App) askForCategory(item *list.Item, index int, count int) (stop bool, err error) {
    fmt.Println("=====================================")
    fmt.Printf("Name:\n    %s\n", item.Name)
    fmt.Printf("URL:\n    %s\n", item.Link)
    fmt.Printf("Description:\n    %s\n", item.Description)
    fmt.Printf("AIDescription:\n    %s\n", item.AIDescription)
    fmt.Printf("Position:\n    %d/%d\n", index, count)
    fmt.Println("=====================================")
    categories := append(s.cfg.Root.TitlesTree(0), "Ignore", "Stop")
    var qs = []*survey.Question{
        {
            Name: "Category",
            Prompt: &survey.Select{
                Message:  "Choose a category:",
                Options:  categories,
                Default:  s.cfg.Root.FindTreeForm(item.AICategory),
                PageSize: 30,
                Description: func(value string, index int) string {
                    if value == s.cfg.Root.FindTreeForm(item.AICategory) {
                        return fmt.Sprintf("%d%%", int(item.AICategoryConfidence*100))
                    }
                    return ""
                },
            },
            Validate: survey.Required,
        },
    }
    answer := struct{ Category string }{}
    err = survey.Ask(qs, &answer)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    if answer.Category == "Stop" {
        return true, nil
    }
    if answer.Category == "Ignore" {
        item.Ignore = true
        var qs = []*survey.Question{
            {
                Name: "Reason",
                Prompt: &survey.Input{
                    Message: "Reason:",
                    Default: item.IgnoreReason,
                },
            },
        }
        answer := struct{ Reason string }{}
        err = survey.Ask(qs, &answer)
        if err != nil {
            if err == terminal.InterruptErr {
                os.Exit(0)
            }
        }
        item.IgnoreReason = answer.Reason
        return false, nil
    }
    item.Category = s.cfg.Root.FindByTree(answer.Category)
    return false, nil
}

func (s *App) confirmDataReplacement() bool {
    var qs = []*survey.Question{
        {
            Name: "Replace",
            Prompt: &survey.Confirm{
                Message: "Replace data with temp data?",
                Default: true,
            },
        },
    }
    answer := struct{ Replace bool }{}
    err := survey.Ask(qs, &answer)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    return answer.Replace
}

func (s *App) confirmGenerateReadme() bool {
    var qs = []*survey.Question{
        {
            Name: "Generate",
            Prompt: &survey.Confirm{
                Message: "Generate readme?",
                Default: true,
            },
        },
    }
    answer := struct{ Generate bool }{}
    err := survey.Ask(qs, &answer)
    if err != nil {
        if err == terminal.InterruptErr {
            os.Exit(0)
        }
    }
    return answer.Generate
}
