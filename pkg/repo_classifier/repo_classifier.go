package repo_classifier

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/korchasa/awesome-toolkit/pkg/config"
    "github.com/korchasa/awesome-toolkit/pkg/list"
    "github.com/sashabaranov/go-openai"
    log "github.com/sirupsen/logrus"
    "strings"
)

const readmeLimit = 4000
const nonEnglishDescriptionPrompt = "repository with non english description"

type RepoClassifier struct {
    aiClient        *openai.Client
    requestTemplate openai.ChatCompletionRequest
    catsTree        *config.CategoryDescription
}

func NewRepoClassifier(aiClient *openai.Client, rootCategory *config.CategoryDescription) *RepoClassifier {
    return &RepoClassifier{
        aiClient:        aiClient,
        requestTemplate: buildRequestTemplate(rootCategory),
        catsTree:        rootCategory,
    }
}

func (r *RepoClassifier) ClassifyRepo(ctx context.Context, item *list.Item, readmeContent string) error {
    req := r.requestTemplate
    txt := fmt.Sprintf(
        "Name:%s\nLink:%s\nLanguage:%s\n%s\n\n%s",
        item.Name, item.Link, item.Language, item.Description, limitString(readmeContent, readmeLimit),
    )
    req.Messages = append(req.Messages, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: txt,
    })

    resp, err := r.aiClient.CreateChatCompletion(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to create openai chat completion: %w", err)
    }

    type choiceStruct struct {
        Category   string  `json:"category"`
        Confidence float32 `json:"confidence"`
        Info       string  `json:"info"`
    }
    var choice choiceStruct
    log.Debugf("chatGPT response: %+v", resp.Choices[0].Message.Content)
    err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &choice)
    if err != nil {
        return fmt.Errorf("failed to parse response: %w: %s", err, resp.Choices[0].Message.Content)
    }
    item.AICategory = r.catsTree.FindTitleByPrompt(strings.Trim(choice.Category, " "))
    item.AICategoryConfidence = choice.Confidence
    item.AIDescription = choice.Info
    return nil
}

func buildRequestTemplate(root *config.CategoryDescription) openai.ChatCompletionRequest {
    prompt := `
I want you to act as a it specialist. I will give you a information about the github repository, and you must answer me only in JSON format, without any explanations. Response JSON format schema:
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "category": {
      "type": "string",
      "enum": %%categories%%,
      "description": "Category name"
    },
    "confidence": {
      "type": "number",
      "minimum": 0,
      "maximum": 1,
      "description": "Confidence in the correctness of the category name"
    },
    "info": {
      "type": "string",
      "description": "Repository description in one paragraph, translated to English"
    }
  },
  "required": [
    "category",
    "confidence",
    "info"
  ]
}
`
    cats := append(root.Prompts(), nonEnglishDescriptionPrompt)
    js, _ := json.MarshalIndent(cats, "", "  ")
    prompt = strings.Replace(prompt, "%%categories%%", string(js), 1)

    return openai.ChatCompletionRequest{
        Model: openai.GPT3Dot5Turbo,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: prompt,
            },
        },
    }
}

func limitString(s string, limit int) string {
    if len(s) <= limit {
        return s
    }
    return s[:limit]
}
