package main

import (
    "context"
    "fmt"
    "github.com/google/go-github/v52/github"
    "github.com/korchasa/awesome-toolkit/pkg/awesome_list"
    "github.com/korchasa/awesome-toolkit/pkg/service"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v3"
    "os"
)

func init() {
    log.SetOutput(os.Stderr)
    log.SetReportCaller(false)
    log.SetLevel(log.InfoLevel)
    log.SetFormatter(
        &log.TextFormatter{
            ForceColors: true,
        },
    )
}

func main() {
    ctx := context.Background()

    data, err := loadData(os.Args[1])
    if err != nil {
        log.Fatalf("failed to load config: %s", err)
    }
    gh := github.NewTokenClient(ctx, ensureEnv("AWESOME_GITHUB_TOKEN"))
    ai := openai.NewClient(ensureEnv("OPENAI_API_KEY"))

    srv, err := service.NewService(gh, ai)
    if err != nil {
        log.Fatalf("failed to create service: %s", err)
    }
    err = srv.Run(ctx, data)
    if err != nil {
        log.Fatalf("failed to run service: %s", err)
    }
    b, err := yaml.Marshal(data)
    if err != nil {
        log.Fatalf("failed to marshal data: %s", err)
    }
    err = os.WriteFile("./data.yaml", b, 0644)
    if err != nil {
        log.Fatalf("failed to write data: %s", err)
    }
}

func loadData(path string) (*awesome_list.List, error) {
    bt, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    cfg := awesome_list.List{}
    err = yaml.Unmarshal(bt, &cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    return &cfg, nil
}

func ensureEnv(name string) string {
    value := os.Getenv(name)
    if value == "" {
        panic(fmt.Errorf("%s env is empty", name))
    }
    return value
}
