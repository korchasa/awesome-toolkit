package main

import (
    "context"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/commands/adder"
    "github.com/korchasa/awesome-toolkit/pkg/commands/collector"
    "github.com/korchasa/awesome-toolkit/pkg/commands/readme"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    log "github.com/sirupsen/logrus"
    "os"
)

const (
    CommandAdd     = "add"
    CommandCollect = "collect"
    CommandReadme  = "readme"
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

type Command interface {
    Run(ctx context.Context) error
}

func main() {
    openaiToken := ensureEnv("OPENAI_API_KEY")
    githubToken := ensureEnv("AWESOME_GITHUB_TOKEN")

    ctx := context.Background()

    if len(os.Args) < 3 {
        log.Fatalf("Usage: %s <work dir> <command>", os.Args[0])
    }

    cfg, err := config.NewFromDir(os.Args[1])
    if err != nil {
        log.Fatalf("failed to load config: %s", err)
    }

    var cmd Command
    switch os.Args[2] {
    case CommandAdd:
        cmd = adder.MustBuildApp(openaiToken, cfg)
    case CommandCollect:
        cmd = collector.MustBuildApp(githubToken, openaiToken, cfg)
    case CommandReadme:
        cmd = readme.MustBuildApp(cfg)
    default:
        log.Fatalf("unknown command: %s", os.Args[2])
    }

    log.Infof("Starting application")
    err = cmd.Run(ctx)
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
