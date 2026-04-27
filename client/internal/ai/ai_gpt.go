package ai

import (
	"context"
	"os"

	"github.com/sashabaranov/go-openai"
)

// TODO fix this system prompt, allow user too configure it at rutntime
const SystemPrompt = `You are an expert at generating Windows commands.
Respond with ONLY the raw command. No markdown, no code fences, no explanation.
GOOD: whoami /all
BAD: Here is your command: whoami /all
BAD: ` + "```" + `powershell
whoami /all
` + "```" + `
Match the execution environment specified in the request (powershell or cmd).`

type GPTProvider struct {
	apiKey string
	model  string
}

func NewGPTProvider() *GPTProvider {
	return &GPTProvider{
		apiKey: os.Getenv("OPENAI_KEY"),
		model:  "gpt-4.1-nano",
	}
}
func (g *GPTProvider) Name() string      { return "gpt" }
func (g *GPTProvider) Model() string     { return g.model }
func (g *GPTProvider) SetModel(m string) { g.model = m }
func (g *GPTProvider) APIKeySet() bool   { return g.apiKey != "" }

func (g *GPTProvider) Generate(prompt string, temp float32) (string, error) {
	return g.OpenAi_CmdGenerate(g.apiKey, g.model, prompt, temp)
}

func (g *GPTProvider) OpenAi_CmdGenerate(apiKey, mode, prompt string, temp float32) (string, error) {
	client := openai.NewClient(apiKey)

	//start := time.Now()
	resp, err := client.CreateChatCompletion(
		context.Background(),

		openai.ChatCompletionRequest{
			Model:       mode,
			Temperature: temp,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	//a.UI.PrintTitle(fmt.Sprintf("AI request completed in %s", time.Since(start).Round(time.Millisecond)))
	return resp.Choices[0].Message.Content, nil

}
