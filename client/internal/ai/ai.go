package ai

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/pflag"
	"github.com/z3vxo/kronos/internal/ui"
)

type AI struct {
	UI        *ui.UI
	Providers map[string]Provider
	active    string
	temp      float32
}

func NewAI(u *ui.UI) (*AI, error) {
	providers := map[string]Provider{
		"claude": NewClaudeProvider(),
		"gpt":    NewGPTProvider(),
	}
	return &AI{
		UI:        u,
		Providers: providers,
		active:    "claude",
		temp:      0.3,
	}, nil
}

func (a *AI) current() (Provider, error) {
	p, ok := a.Providers[a.active]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %q", a.active)
	}
	return p, nil
}

func (a *AI) HandleAI(args []string) {
	if len(args) == 0 {
		a.UI.Send(ui.BAD.Sprint("Usage ai <args>"))
		return
	}

	sub := map[string]func([]string) error{
		"reconfig": a.HandleAiReconfig,
		"settings": func(_ []string) error { a.PrintSettings(); return nil },
		"models":   func(_ []string) error { a.PrintModels(); return nil },
	}
	if fn, ok := sub[args[0]]; ok {
		if err := fn(args); err != nil {
			a.UI.Send(ui.BAD.Sprintf("Error: %s", err))
		}
		return
	}

	p, err := a.current()
	if err != nil {
		a.UI.Send(ui.BAD.Sprintf("Error: %s", err))
		return
	}
	if !p.APIKeySet() {
		a.UI.Send(ui.BAD.Sprintf("API Key Not set for provider %q", p.Name()))
		return
	}

	resp, err := p.Generate(strings.Join(args, " "), a.temp)
	if err != nil {
		a.UI.Send(ui.BAD.Sprintf("Error: %s", err))
		return
	}
	a.PrintResponse(p, resp)
}

func (a *AI) PrintResponse(p Provider, resp string) {
	a.UI.Send(ui.INFO.Sprintf("Output from %s[%s]\n---------------\n%s", p.Name(), p.Model(), resp))
	a.UI.Send("---------------\n")
}

func (a *AI) PrintSettings() {
	p, err := a.current()
	if err != nil {
		a.UI.Send(ui.BAD.Sprintf("error: %s", err))
		return
	}
	apiSet := "FALSE"
	if p.APIKeySet() {
		apiSet = "TRUE"
	}
	a.UI.Send(ui.INFO.Sprintf("Provider: %s", p.Name()))
	a.UI.Send(ui.INFO.Sprintf("Model: %s", p.Model()))
	a.UI.Send(ui.INFO.Sprintf("Temp: %f", a.temp))
	a.UI.Send(ui.INFO.Sprintf("API Key Set: %s", apiSet))
}

func (a *AI) HandleAiReconfig(args []string) error {
	fs := pflag.NewFlagSet("start", pflag.ContinueOnError)
	provider := fs.StringP("provider", "p", "", "")
	model := fs.StringP("model", "m", "", "")
	temp := fs.Float32P("temperature", "t", a.temp, "")
	if err := fs.Parse(args[1:]); err != nil {
		a.UI.Send(ui.WARN.Sprintf("[!] %v", err))
		return err
	}
	if *provider == "" || *model == "" {
		return errors.New("missing provider or model arguments")
	}

	p, ok := a.Providers[*provider]
	if !ok {
		return fmt.Errorf("unknown provider %q (avaliable %s)", *provider, a.providerNames)
	}

	p.SetModel(*model)
	a.active = *provider
	a.temp = *temp
	a.UI.PrintTitle(fmt.Sprintf("Using provider %s with model %s (temp %.1f)", *provider, *model, *temp))
	if !p.APIKeySet() {
		a.UI.Send(ui.WARN.Sprintf("Api Key not set for %s, ensure to set it in .bashrc/.zshrc before use", *provider))
	}

	return nil
}

func (a *AI) providerNames() string {
	names := make([]string, 0, len(a.Providers))
	for n := range a.Providers {
		names = append(names, n)
	}
	return strings.Join(names, ":")
}

var modelCatalog = map[string][]string{
	"CLAUDE": {
		"claude-opus-4-6",
		"claude-opus-4.5",
		"claude-sonnet-4.6",
		"claude-sonnet-4.5",
		"claude-haiku-4-5-20251001",
		"claude-haiku-4-5",
	},
	"GPT": {
		"gpt-5",
		"gpt-5-mini",
		"gpt-5-nano",
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		"gpt-4o",
		"gpt-4o-mini",
		"o3",
		"o3-mini",
	},
}

var catalogOrder = []string{"CLAUDE", "GPT"}

func (a *AI) PrintModels() {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)

	header := make(table.Row, len(catalogOrder))
	configs := make([]table.ColumnConfig, len(catalogOrder))
	for i, group := range catalogOrder {
		header[i] = group
		configs[i] = table.ColumnConfig{Number: i + 1, AlignHeader: text.AlignCenter, Align: text.AlignCenter}
	}
	t.AppendHeader(header)
	t.SetColumnConfigs(configs)

	maxRows := 0
	for _, group := range catalogOrder {
		if l := len(modelCatalog[group]); l > maxRows {
			maxRows = l
		}
	}

	for i := 0; i < maxRows; i++ {
		row := make(table.Row, len(catalogOrder))
		for j, group := range catalogOrder {
			if models := modelCatalog[group]; i < len(models) {
				row[j] = models[i]
			}
		}
		t.AppendRow(row)
	}

	a.UI.Send(t.Render())
}
