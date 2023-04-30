package main

import (
    "context"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/service"
    log "github.com/sirupsen/logrus"
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

    cfg, err := config.NewFromFile(os.Args[1])
    if err != nil {
        log.Fatalf("failed to load config: %s", err)
    }

    srv, err := service.BuildService(
        ensureEnv("AWESOME_GITHUB_TOKEN"),
        ensureEnv("OPENAI_API_KEY"),
        cfg)
    if err != nil {
        log.Fatalf("failed to create service: %s", err)
    }

    log.Infof("Starting service")
    err = srv.Run(ctx)
    if err != nil {
        log.Fatalf("failed to run service: %s", err)
    }

    //readme, err := readme_generator.NewReadmeGenerator().Generate(data)
    //if err != nil {
    //    log.Fatalf("failed to generate readme: %s", err)
    //}
    //err = os.WriteFile("./README.md", []byte(readme), 0644)
}

func ensureEnv(name string) string {
    value := os.Getenv(name)
    if value == "" {
        panic(fmt.Errorf("%s env is empty", name))
    }
    return value
}
