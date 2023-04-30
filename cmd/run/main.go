package main

import (
    "context"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/app"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    log "github.com/sirupsen/logrus"
    "os"
)

func init() {
    log.SetOutput(os.Stderr)
    log.SetReportCaller(false)
    log.SetLevel(log.DebugLevel)
    log.SetFormatter(
        &log.TextFormatter{
            ForceColors: true,
        },
    )
}

func main() {
    ctx := context.Background()

    cfg, err := config.NewFromFile(os.Args[1])
    if err != nil {
        log.Fatalf("failed to load config: %s", err)
    }

    srv := app.MustBuildApp(
        ensureEnv("AWESOME_GITHUB_TOKEN"),
        ensureEnv("OPENAI_API_KEY"),
        cfg)
    if err != nil {
        log.Fatalf("failed to create service: %s", err)
    }

    log.Infof("Starting application")
    err = srv.Run(ctx)
    if err != nil {
        log.Fatalf("failed to run service: %s", err)
    }
}

func ensureEnv(name string) string {
    value := os.Getenv(name)
    if value == "" {
        panic(fmt.Errorf("%s env is empty", name))
    }
    return value
}
