package service

import (
    "context"
    "fmt"
    "github.com/google/go-github/v52/github"
    "github.com/korchasa/awesome-toolkit/pkg/awesome_list"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "strings"
)

const readmeLimit = 10000
const nonEnglishDescriptionPrompt = "repository with non english description"

type Service struct {
    gh  *github.Client
    gpt *openai.Client
}

func NewService(gh *github.Client, aiToken *openai.Client) (*Service, error) {
    return &Service{
        gh:  gh,
        gpt: aiToken,
    }, nil
}

func (s *Service) Run(ctx context.Context, data *awesome_list.List) error {
    opt := &github.SearchOptions{
        Sort:      "stars",
        Order:     "desc",
        TextMatch: true,
        ListOptions: github.ListOptions{
            Page:    0,
            PerPage: 100,
        },
    }

    count := 0
    err := s.processRepositories(ctx, data, opt, &count)
    if err != nil {
        return fmt.Errorf("failed to process repositories: %w", err)
    }
    return nil
}

func (s *Service) processRepositories(ctx context.Context, data *awesome_list.List, opt *github.SearchOptions, count *int) error {
    rps, _, err := s.gh.Search.Repositories(ctx, data.Query, opt)
    if err != nil {
        return fmt.Errorf("failed to search repos: %w", err)
    }
    log.Infof("loaded %d repos", len(rps.Repositories))
    for _, repo := range rps.Repositories {
        *count++
        log.Infof("%d/%d processing: %s", *count, *rps.Total, *repo.HTMLURL)
        if data.ItemExists(*repo.HTMLURL) {
            log.Infof("%d/%d skipped: %s", *count, *rps.Total, *repo.HTMLURL)
            continue
        }
        err := s.processRepository(ctx, repo, data)
        if err != nil {
            log.WithError(err).Errorf("failed to process %s", *repo.HTMLURL)
        } else {
            log.Infof("%d/%d processed: %s", *count, *rps.Total, *repo.HTMLURL)
        }
    }
    if *count >= *rps.Total {
        return nil
    }
    opt.Page++
    return s.processRepositories(nil, data, opt, count)
}

func (s *Service) processRepository(ctx context.Context, repo *github.Repository, data *awesome_list.List) error {
    categoryPrompt, description, err := s.askRepoCategoryAndDescription(ctx, repo, data)
    if err != nil {
        return fmt.Errorf("failed to generate info for `%s`: %w", *repo.HTMLURL, err)
    }
    if categoryPrompt == nonEnglishDescriptionPrompt {
        return nil
    }
    item := awesome_list.Item{
        Name:          *repo.FullName,
        Link:          *repo.HTMLURL,
        Description:   *repo.Description,
        AIDescription: description,
    }
    if !data.AddItemByTitlePrompt(categoryPrompt, item) {
        return fmt.Errorf("failed to add item `%s` to list, prompt `%s` not found", *repo.FullName, categoryPrompt)
    }
    return nil
}

func (s *Service) askRepoCategoryAndDescription(ctx context.Context, repo *github.Repository, data *awesome_list.List) (string, string, error) {
    readme, _, err := s.gh.Repositories.GetReadme(ctx, *repo.Owner.Login, *repo.Name, nil)
    if err != nil {
        return "", "", fmt.Errorf("failed to get readme: %w", err)
    }
    readmeContent, err := readme.GetContent()
    if err != nil {
        return "", "", fmt.Errorf("failed to get readme content: %w", err)
    }

    resp, err := s.gpt.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT3Dot5Turbo,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: s.buildPrompt(data, *repo.FullName, readmeContent),
            },
        },
    })
    if err != nil {
        return "", "", fmt.Errorf("failed to create completion: %w", err)
    }

    parts := strings.SplitN(resp.Choices[0].Message.Content, "/", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("failed to parse response: %s", resp.Choices[0].Message.Content)
    }
    return strings.Trim(parts[0], " "), strings.Trim(parts[1], " "), nil
}

func (s *Service) buildPrompt(data *awesome_list.List, repoTitle string, repoDesc string) string {
    prompt := "I want you to act as a repository classifier. I will give you a description of the repository, and you must answer me with one of the categories and description in format: <category> / <description>. The format is very important! List of categories:"
    for _, cp := range data.ListCategoriesPrompts() {
        prompt += fmt.Sprintf("- %s\n", cp)
    }
    prompt += fmt.Sprintf("- %s\n", nonEnglishDescriptionPrompt)
    prompt += fmt.Sprintf("Repository: %s\n%s\n", repoTitle, repoDesc)
    if len(prompt) > readmeLimit {
        prompt = prompt[0:readmeLimit]
    }
    return prompt
}
