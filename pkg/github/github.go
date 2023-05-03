package github

import (
    "context"
    "fmt"
    "github.com/google/go-github/v52/github"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    log "github.com/sirupsen/logrus"
    "strings"
    "time"
)

type GitHub struct {
    client *github.Client
}

func NewGitHub(token string) *GitHub {
    return &GitHub{
        client: github.NewTokenClient(context.TODO(), token),
    }
}

func (g *GitHub) SearchRepos(ctx context.Context, query string) ([]*list.Item, error) {
    var result []*list.Item
    err := g.search(ctx, &result, query, 1)
    if err != nil {
        return nil, fmt.Errorf("failed to search: %w", err)
    }
    return result, nil
}

func (g *GitHub) GetReadme(ctx context.Context, item *list.Item) (string, error) {
    parts := strings.Split(item.Name, "/")
    if len(parts) != 2 {
        log.Infof("invalid repo name: %s", item.Name)
        return "", nil
    }
    readme, _, err := g.client.Repositories.GetReadme(ctx, parts[0], parts[1], nil)
    if err != nil {
        return "", fmt.Errorf("failed to get readme: %w", err)
    }
    readmeContent, err := readme.GetContent()
    if err != nil {
        return "", fmt.Errorf("failed to get readme content: %w", err)
    }
    return readmeContent, nil
}

func (g *GitHub) search(ctx context.Context, result *[]*list.Item, query string, page int) error {
    log.Infof("searching page %d", page)
    opt := &github.SearchOptions{
        //Sort:      "stars",
        //Order:     "desc",
        TextMatch: true,
        ListOptions: github.ListOptions{
            Page:    page,
            PerPage: 100,
        },
    }

    rps, _, err := g.client.Search.Repositories(ctx, query, opt)
    if err != nil {
        return fmt.Errorf("failed to search repos: %w", err)
    }
    for _, r := range rps.Repositories {
        *result = append(*result, &list.Item{
            Name:        limitString(safeGet(r.FullName), 500),
            Link:        limitString(safeGet(r.HTMLURL), 500),
            Description: limitString(safeGet(r.Description), 500),
            Language:    limitString(safeGet(r.Language), 500),
            IsNew:       true,
            CreatedAt:   time.Now(),
        })
    }
    log.Infof("found %d repos", len(rps.Repositories))
    if len(rps.Repositories) != 0 {
        return g.search(ctx, result, query, page+1)
    }
    return nil
}

func safeGet[T any](s *T) T {
    var empty T
    if s == nil {
        return empty
    }
    return *s
}

func limitString(s string, limit int) string {
    if len(s) > limit {
        return s[:limit]
    }
    return s
}
