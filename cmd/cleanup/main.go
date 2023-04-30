package main

import (
	"github.com/korchasa/awesome-toolkit/pkg/config"
	"github.com/korchasa/awesome-toolkit/pkg/list"
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
	cfg, err := config.NewFromDir(os.Args[1])
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	data, err := list.NewFromFile(cfg.DataPath())
	if err != nil {
		log.Fatalf("failed to load data: %s", err)
	}

	var newData []*list.Item
	for i, item := range data.Items {
		if item.Ignore && item.IgnoreReason == "non latin description" {
			log.Infof("skipping %d: %s", i, item.Name)
		} else {
			newData = append(newData, item)
		}
	}
	data.Items = newData
	err = data.Save(cfg.DataPath())
	if err != nil {
		log.Fatalf("failed to save data: %s", err)
	}
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
