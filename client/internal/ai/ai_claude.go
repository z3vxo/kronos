package ai

import (
	"context"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	apiKey string
	model  string
}

func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: os.Getenv("CLAUDE_KEY"),
		model:  "claude-haiku-4-5-20251001",
	}
}
func (g *ClaudeProvider) Name() string      { return "claude" }
func (g *ClaudeProvider) Model() string     { return g.model }
func (g *ClaudeProvider) SetModel(m string) { g.model = m }
func (g *ClaudeProvider) APIKeySet() bool   { return g.apiKey != "" }

func (g *ClaudeProvider) Generate(prompt string, temp float32) (string, error) {

	return g.Claude_CmdGenerate(g.apiKey, g.Model(), prompt, temp)
}

func (a *ClaudeProvider) Claude_CmdGenerate(apiKey, mode, prompt string, temp float32) (string, error) {

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	//start := time.Now()
	resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:       mode,
		MaxTokens:   1024,
		Temperature: anthropic.Float(float64(temp)),
		System: []anthropic.TextBlockParam{
			{Text: SystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}

	//a.UI.PrintTitle(fmt.Sprintf("AI request completed in %s", time.Since(start).Round(time.Millisecond)))

	return resp.Content[0].Text, nil
}
