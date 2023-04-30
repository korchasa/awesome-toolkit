package main

import (
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/korchasa/awesome-toolkit/pkg/readme_generator"
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
    cfg, err := config.NewFromFile(os.Args[1])
    if err != nil {
        log.Fatalf("failed to load config: %s", err)
    }

    ll, err := list.NewFromFile(cfg.DataPath)
    if err != nil {
        log.Fatalf("failed to load list: %s", err)
    }

    str, err := readme_generator.NewReadmeGenerator().Generate(cfg, ll)
    if err != nil {
        log.Fatalf("failed to generate readme: %s", err)
    }
    err = os.WriteFile("./README.md", []byte(str), 0644)
}
