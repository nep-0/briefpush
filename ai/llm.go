package ai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

const (
	SUMMARY_PROMPT = "You are a helpful assistant. Please summarize the following text in original language:"
	REPORT_PROMPT  = "You are a helpful assistant. Please create a brief report based on the following summaries in original language:"
)

type LLM struct {
	BaseURL string
	ApiKey  string
	Model   string

	client *openai.Client

	summaryPrompt string
	reportPrompt  string
}

func NewLLM(baseURL string, apiKey string, model string) *LLM {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)
	return &LLM{
		BaseURL:       baseURL,
		ApiKey:        apiKey,
		Model:         model,
		client:        client,
		summaryPrompt: SUMMARY_PROMPT,
		reportPrompt:  REPORT_PROMPT,
	}
}

func (l *LLM) GenerateSummary(ctx context.Context, title string, content string) (string, error) {
	resp, err := l.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: l.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: l.summaryPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: title + "\n\n" + content,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (l *LLM) GenerateReport(ctx context.Context, summaries []string) (string, error) {
	messages := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: l.reportPrompt}}
	for _, summary := range summaries {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: summary,
		})
	}
	resp, err := l.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    l.Model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
