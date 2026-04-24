package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func (a *AI) Claude_CmdGenerate(msg string) (string, error) {
	if a.ApiKey["claude"] == "" {
		return "", errors.New("Missing API Key")
	}
	client := anthropic.NewClient(
		option.WithAPIKey(a.ApiKey["claude"]),
	)
	start := time.Now()
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:       a.Model["claude"],
		MaxTokens:   1024,
		Temperature: anthropic.Float(float64(a.Temp)),
		System: []anthropic.TextBlockParam{
			{Text: SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(msg)),
		},
	})
	if err != nil {
		return "", err
	}

	a.UI.PrintTitle(fmt.Sprintf("AI request completed in %s", time.Since(start).Round(time.Millisecond)))

	return resp.Content[0].Text, nil
}
