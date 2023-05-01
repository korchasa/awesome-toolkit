package cleanup

import (
	"context"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/korchasa/awesome-toolkit/pkg/config"
	"github.com/korchasa/awesome-toolkit/pkg/list"
	log "github.com/sirupsen/logrus"
	"os"
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
		log.Fatalf("failed to load data: %s", err)
	}
	return data
}

func (s *App) Run(_ context.Context) error {
	log.Infof("Initial items count: %d", len(s.data.Items))

	s.removeDuplicates(s.data)
	s.removeItemsWithoutCategory(s.data)

	if askForSave() {
		err := s.data.Save(s.cfg.DataPath())
		if err != nil {
			log.Fatalf("failed to save data: %s", err)
		}
	}
	return nil
}

func (s *App) removeDuplicates(data *list.List) {
	log.Infof("Remove duplicates...")
	uniqueMap := make(map[string]*list.Item)
	var result []*list.Item

	for _, elem := range data.Items {
		if _, found := uniqueMap[elem.Link]; !found {
			uniqueMap[elem.Link] = elem
			result = append(result, elem)
		} else {
			log.Printf("Removed duplicate: %s", elem.Link)
		}
	}
	log.Infof("Removed %d duplicates", len(data.Items)-len(result))
	data.Items = result
}

func (s *App) removeItemsWithoutCategory(data *list.List) {
	log.Infof("Remove items without category...")
	var newItems []*list.Item
	for _, item := range data.Items {
		if item.Category == "" && !item.Ignore {
			log.Infof("Removed item without category: %s", item.Link)
			continue
		}
		newItems = append(newItems, item)
	}
	log.Infof("Removed %d items", len(data.Items)-len(newItems))
	data.Items = newItems
}

func askForSave() bool {
	var qs = []*survey.Question{
		{
			Name: "Confirm",
			Prompt: &survey.Confirm{
				Message: "Save?",
				Default: true,
			},
		},
	}
	answer := struct{ Confirm bool }{}
	err := survey.Ask(qs, &answer)
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(0)
		}
	}
	return answer.Confirm
}
