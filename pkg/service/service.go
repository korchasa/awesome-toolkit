package service

import (
    "context"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/github"
    "github.com/korchasa/awesome-toolkit/pkg/ignorer"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/repo_classifier"
    log "github.com/sirupsen/logrus"
    "os"
)

type Service struct {
    github     *github.GitHub
    classifier *repo_classifier.RepoClassifier
    cfg        *config.Config
    data       *list.List
    ignorer    *ignorer.Ignorer
}

func BuildService(ghToken string, aiToken string, cfg *config.Config) (*Service, error) {
    data, err := list.NewFromFile(cfg.DataPath)
    if os.IsNotExist(err) {
        log.Infof("data file `%s` not found, creating new one", cfg.DataPath)
        data = list.NewEmpty(cfg.Title)
    } else if err != nil {
        return nil, fmt.Errorf("failed to load data: %w", err)
    }

    return &Service{
        github:     github.NewGitHub(ghToken),
        classifier: repo_classifier.NewRepoClassifier(aiToken, cfg),
        cfg:        cfg,
        data:       data,
        ignorer:    ignorer.NewIgnorer(),
    }, nil
}

func (s *Service) Run(ctx context.Context) error {
    items, err := s.github.SearchRepos(ctx, s.cfg.Query)
    if err != nil {
        return fmt.Errorf("failed to search repositories: %w", err)
    }
    for _, item := range items {
        err := s.processFoundRepo(ctx, item)
        if err != nil {
            log.WithError(err).Warnf("failed to process repo `%s`", item.Name)
            continue
        }
        s.data.Add(item)
    }
    log.Infof("saving data to `%s`", s.cfg.DataPath)
    err = s.data.Save(s.cfg.DataPath)
    if err != nil {
        return fmt.Errorf("failed to save data: %w", err)
    }
    return nil
}

func (s *Service) processFoundRepo(ctx context.Context, item *list.Item) error {
    if s.data.ItemExists(item) {
        log.Infof("Skip `%s` because it already exists in data", item.Name)
        return nil
    }
    s.ignorer.ResolveIgnores(item)
    readme, err := s.github.GetReadme(ctx, item)
    if err != nil {
        return fmt.Errorf("failed to get readme for `%s`: %w", item.Name, err)
    }
    err = s.classifier.ClassifyRepo(ctx, item, readme)
    if err != nil {
        return fmt.Errorf("failed to classify repo `%s`: %w", item.Name, err)
    }
    log.Infof("Processed %s", item)
    return nil
}
