package collector

import (
    "context"
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/AlecAivazis/survey/v2/terminal"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/github"
    "github.com/korchasa/awesome-toolkit/pkg/ignorer"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/repo_classifier"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "os"
    "time"
)

type App struct {
    github       *github.GitHub
    classifier   *repo_classifier.RepoClassifier
    tempData     *list.List
    ignorer      *ignorer.Ignorer
    categoryTree *config.CategoryDescription
    query        string
    dataPath     string
    tempDataPath string
}

func MustBuildApp(gh *github.GitHub, ai *openai.Client, cfg *config.Config) *App {
    return &App{
        github:       gh,
        classifier:   repo_classifier.NewRepoClassifier(ai, cfg.Root),
        tempData:     mustLoadTempData(cfg),
        ignorer:      ignorer.NewIgnorer(),
        categoryTree: cfg.Root,
        dataPath:     cfg.DataPath(),
        tempDataPath: cfg.TempDataPath(),
        query:        cfg.Query,
    }
}

func mustLoadTempData(cfg *config.Config) *list.List {
    tempData, err := list.NewFromFile(cfg.TempDataPath())
    if os.IsNotExist(err) {
        log.Infof("temp data file `%s` not found, creating new one", cfg.TempDataPath())
        oldData, err := list.NewFromFile(cfg.DataPath())
        if os.IsNotExist(err) {
            log.Infof("data file `%s` not found, creating new one", cfg.DataPath())
            oldData = list.NewEmpty()
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
    items, err := s.findNewRepos(ctx)
    if err != nil {
        return fmt.Errorf("failed to find new repos: %w", err)
    }
    count := len(items)
    for i, item := range items {
        stop, err := s.processFoundRepo(ctx, item, i, count)
        if err != nil {
            log.Errorf("failed to process repo `%s`: %s", item.Name, err)
            continue
        }
        if stop {
            break
        }
        s.tempData.Add(item)
        if err := s.tempData.Save(s.tempDataPath); err != nil {
            return fmt.Errorf("failed to save temp data: %w", err)
        }
    }
    if s.confirmDataReplacement() {
        log.Infof("saving data to `%s`", s.dataPath)
        s.tempData.UpdatedAt = time.Now()
        err = s.tempData.Save(s.dataPath)
        if err != nil {
            return fmt.Errorf("failed to save data: %w", err)
        }
        if err := os.Remove(s.tempDataPath); err != nil {
            return fmt.Errorf("failed to remove temp data: %w", err)
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
    fmt.Printf("Language:\n    %s\n", item.Language)
    fmt.Printf("Position:\n    %d/%d\n", index, count)
    fmt.Println("=====================================")
    categories := append(s.categoryTree.TitlesTree(0), "Ignore", "Stop")
    var qs = []*survey.Question{
        {
            Name: "Category",
            Prompt: &survey.Select{
                Message:  "Choose a category:",
                Options:  categories,
                Default:  s.categoryTree.FindTreeForm(item.AICategory),
                PageSize: 20,
                Description: func(value string, index int) string {
                    if value == s.categoryTree.FindTreeForm(item.AICategory) {
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
    item.Category = s.categoryTree.FindByTree(answer.Category)
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

func (s *App) findNewRepos(ctx context.Context) ([]*list.Item, error) {
    items, err := s.github.SearchRepos(ctx, s.query)
    if err != nil {
        return nil, fmt.Errorf("failed to search repositories: %w", err)
    }
    var newItems []*list.Item
    for _, item := range items {
        if !s.tempData.ItemExists(item) {
            newItems = append(newItems, item)
        }
    }
    return newItems, nil
}
