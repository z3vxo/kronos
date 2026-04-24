package ai

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/pflag"
	"github.com/z3vxo/kronos/internal/ui"
)

type AI struct {
	UI       *ui.UI
	ApiKey   map[string]string
	Provider string
	Model    map[string]string
	Temp     float32
}

func NewAI(u *ui.UI) (*AI, error) {
	gptAPIkey := os.Getenv("OPENAI_KEY")
	CladeAPIkey := os.Getenv("CLAUDE_KEY")
	return &AI{
		UI:       u,
		ApiKey:   map[string]string{"claude": CladeAPIkey, "gpt": gptAPIkey},
		Model:    map[string]string{"claude": "claude-haiku-4-5-20251001", "gpt": "gpt-4.1-nano"},
		Provider: "claude",
		Temp:     0.3,
	}, nil
}

func (a *AI) HandleAI(args []string) {
	if len(args) == 0 {
		a.UI.Send(ui.BAD.Sprint("Usage: ai <args>"))
		return
	}

	switch args[0] {
	case "reconfig":
		if err := a.HandleAiReconfig(args); err != nil {
			a.UI.Send(err.Error())
			return
		}
		return
	case "settings":
		a.PrintSettings()
		return
	case "models":
		a.PrintModels()
		return
	}

	if a.Provider == "gpt" {
		resp, err := a.OpenAi_CmdGenerate(strings.Join(args, " "))
		if err != nil {
			a.UI.Send(ui.BAD.Sprintf("Error: %s", err))
			return
		}
		a.UI.Send(ui.INFO.Sprintf("Output from gpt[%s]\n------------------------------\n%s", a.Model["gpt"], resp))
		a.UI.Send("------------------------------\n")
		return
	}
	if a.Provider == "claude" {
		resp, err := a.Claude_CmdGenerate(strings.Join(args, " "))
		if err != nil {
			a.UI.Send(ui.BAD.Sprintf("Error: %s", err))
			return
		}
		a.UI.Send(ui.INFO.Sprintf("Output from claude[%s]\n------------------------------\n%s", a.Model["claude"], resp))
		a.UI.Send("------------------------------\n")
		return
	}

}

func (a *AI) PrintSettings() {
	a.UI.Send(ui.INFO.Sprintf("Provider: %s", a.Provider))
	a.UI.Send(ui.INFO.Sprintf("Model: %s", a.Model[a.Provider]))
	a.UI.Send(ui.INFO.Sprintf("Temperature: %.1f", a.Temp))
	apiSet := "TRUE"
	if a.ApiKey[a.Provider] == "" {
		apiSet = "FALSE"
	}
	a.UI.Send(ui.INFO.Sprintf("API Key Set: %s", apiSet))
}

func (a *AI) HandleAiReconfig(args []string) error {
	fs := pflag.NewFlagSet("start", pflag.ContinueOnError)
	provider := fs.StringP("provider", "p", "", "")
	model := fs.StringP("model", "m", "", "")
	temp := fs.Float32P("temperature", "t", a.Temp, "")
	if err := fs.Parse(args[1:]); err != nil {
		a.UI.Send(ui.WARN.Sprintf("[!] %v", err))
		return err
	}
	if *provider == "" || *model == "" {
		return errors.New("missing provider or model arguments")
	}

	a.Provider = *provider
	a.Model[a.Provider] = *model
	a.Temp = *temp
	a.UI.PrintTitle(fmt.Sprintf("Using provider %s with model %s (temp %.1f)", *provider, *model, *temp))
	return nil
}

func (a *AI) PrintModels() {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter},
		{Number: 2, WidthMin: 38, AlignHeader: text.AlignCenter},
	})

	t.AppendHeader(table.Row{"MODEL", "COST"})

	t.AppendRow(table.Row{"CLAUDE", "CLAUDE"}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignCenter})
	t.AppendSeparator()

	t.AppendRows([]table.Row{
		{"claude-opus-4-6", "IN: $5.00/Mtok    | OUT: $25.00/Mtok"},
		{"claude-opus-4-5", "IN: $5.00/Mtok    | OUT: $25.00/Mtok"},
		{"claude-sonnet-4-6", "IN: $3.00/Mtok    | OUT: $15.00/Mtok"},
		{"claude-sonnet-4-5", "IN: $3.00/Mtok    | OUT: $15.00/Mtok"},
		{"claude-haiku-4-5-20251001", "IN: $1.00/Mtok    | OUT: $5.00/Mtok"},
		{"claude-haiku-4-5", "IN: $1.00/Mtok    | OUT: $5.00/Mtok"},
	})

	t.AppendSeparator()

	t.AppendRow(table.Row{"GPT", "GPT"}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignCenter})
	t.AppendSeparator()

	t.AppendRows([]table.Row{
		{"gpt-5", "IN: $1.25/Mtok    | OUT: $10.00/Mtok"},
		{"gpt-5-mini", "IN: $0.25/Mtok    | OUT: $2.00/Mtok"},
		{"gpt-5-nano", "IN: $0.05/Mtok    | OUT: $0.40/Mtok"},
		{"gpt-4.1", "IN: $2.00/Mtok    | OUT: $8.00/Mtok"},
		{"gpt-4.1-mini", "IN: $0.40/Mtok    | OUT: $1.60/Mtok"},
		{"gpt-4.1-nano", "IN: $0.10/Mtok    | OUT: $0.40/Mtok"},
		{"gpt-4o", "IN: $2.50/Mtok    | OUT: $10.00/Mtok"},
		{"gpt-4o-mini", "IN: $0.15/Mtok    | OUT: $0.60/Mtok"},
		{"o3", "IN: $2.00/Mtok    | OUT: $8.00/Mtok"},
		{"o3-mini", "IN: $1.10/Mtok    | OUT: $4.40/Mtok"},
	})

	a.UI.Send(t.Render())
}
