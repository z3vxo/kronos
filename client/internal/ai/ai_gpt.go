package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const SystemPrompt = `You are an expert at generating Windows commands.
Respond with ONLY the raw command. No markdown, no code fences, no explanation.
GOOD: whoami /all
BAD: Here is your command: whoami /all
BAD: ` + "```" + `powershell
whoami /all
` + "```" + `
Match the execution environment specified in the request (powershell or cmd).`

func (a *AI) OpenAi_CmdGenerate(req string) (string, error) {
	if a.ApiKey["gpt"] == "" {
		return "", errors.New("Missing API Key")
	}
	client := openai.NewClient(a.ApiKey["gpt"])

	start := time.Now()
	resp, err := client.CreateChatCompletion(
		context.Background(),

		openai.ChatCompletionRequest{
			Model:       a.Model["gpt"],
			Temperature: a.Temp,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	a.UI.PrintTitle(fmt.Sprintf("AI request completed in %s", time.Since(start).Round(time.Millisecond)))
	return resp.Choices[0].Message.Content, nil

}
