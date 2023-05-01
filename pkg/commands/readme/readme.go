package readme

import (
    "context"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/readme_generator"
    log "github.com/sirupsen/logrus"
    "os"
    "time"
)

type App struct {
    cfg  *config.Config
    data *list.List
}

func MustBuildApp(cfg *config.Config) *App {
    return &App{
        cfg:  cfg,
        data: mustLoadData(cfg),
    }
}

func mustLoadData(cfg *config.Config) *list.List {
    data, err := list.NewFromFile(cfg.DataPath())
    if os.IsNotExist(err) {
        log.Infof("data file `%s` not found, creating new one", cfg.DataPath())
        data = list.NewEmpty()
    } else if err != nil {
        log.Fatalf("failed to load data: %s", err)
    }
    return data
}

func (s *App) Run(_ context.Context) error {
    data, err := list.NewFromFile(s.cfg.DataPath())
    if err != nil {
        log.Fatalf("failed to load old data: %s", err)
    }
    readme, err := readme_generator.NewReadmeGenerator().Generate(s.cfg, data)
    if err != nil {
        log.Fatalf("failed to generate readme: %s", err)
    }
    err = os.WriteFile(s.cfg.ReadmePath(), []byte(readme), 0644)
    if err != nil {
        log.Fatalf("failed to write readme: %s", err)
    }
    data.ReadmeGeneratedAt = time.Now()
    err = data.Save(s.cfg.DataPath())
    if err != nil {
        return fmt.Errorf("failed to save data: %w", err)
    }
    log.Infof("Readme generated at %s", data.ReadmeGeneratedAt)
    return nil
}
