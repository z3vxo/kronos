package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/pflag"
	"github.com/z3vxo/kronos/internal/ui"
)

func (c *CLI) HandleProfiles(args []string) {
	if len(args) < 1 {
		c.ui.Send(ui.BAD.Sprint("Usage: profiles new -f /path/to/file(optional) | delete <name> | list"))
		return
	}

	switch args[0] {
	case "new":
		c.LoadAndSendProfile(args[1:])
		return
	case "list":
		c.ListProfiles()
	case "delete":
		c.DeleteProfile(args[1])
	case "generate":
		c.GenerateAgent(args[1:])
	default:
		c.ui.Send(ui.WARN.Sprint("Missing sub argument"))
	}

	
}

func (c *CLI) GenerateAgent(args []string) {
	fs := pflag.NewFlagSet("generate", pflag.ContinueOnError)
	debug := fs.BoolP("debug", "d", false, "")
	format := fs.StringP("format", "f", "exe", "")
	name := fs.StringP("name", "n", "hades", "")
	profile := fs.StringP("profile", "p", "", "")

	if err := fs.Parse(args); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error parsing flags: %v", err))
		return
	}

	if *profile == "" {
		c.ui.Send(ui.BAD.Sprint("Error: profile must be set"))
		return
	}

	c.ui.PrintTitle("Agent Details")
	c.ui.Send(ui.INFO.Sprintf("Profile: %s", *profile))
	c.ui.Send(ui.INFO.Sprintf("Format:  %s", *format))
	c.ui.Send(ui.INFO.Sprintf("Debug:   %t", *debug))
	c.ui.Send(ui.INFO.Sprintf("Name:    %s", *name))

	req := GeneratePayloadReq{
		Format: *format,
		Debug:  *debug,
		Name:   *name,
	}

	data, err := json.Marshal(&req)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed marshaling json: %s", err))
		return
	}

	resp, err := c.http.DoPostRaw(fmt.Sprintf("ts/rest/agent/generate/%s", *profile), data)
	if err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed generating payload: %s", err))
		return
	}
	defer resp.Body.Close()

	home, _ := os.UserHomeDir()
	buildDir := filepath.Join(home, ".kronos", "builds")
	os.MkdirAll(buildDir, 0700)
	outPath := filepath.Join(buildDir, fmt.Sprintf("%s.%s", *name, *format))

	f, err := os.Create(outPath)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed creating output file: %s", err))
		return
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed writing payload: %s", err))
		return
	}

	size := fmt.Sprintf("%.1fkb", float64(n)/1024)
	c.ui.PrintTitle(fmt.Sprintf("Payload saved to %s [%s]", outPath, size))
}


func (c *CLI) DeleteProfile(name string) {

	if err := c.http.DoPost(fmt.Sprintf("ts/rest/agent/profiles/delete/%s", name), nil, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed deleting profile: %v", err))
		return
	}

	c.ui.PrintTitle(fmt.Sprint("Profile Deleted!"))
}

func (c *CLI) ListProfiles() {
	var profiles []Profile
	if err := c.http.DoGet("ts/rest/agent/profiles/list", &profiles); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed listing profiles: %v", err))
		return
	}

	if len(profiles) == 0 {
		c.ui.Send(ui.INFO.Sprint("No profiles found"))
		return
	}

	c.ui.PrintTitle("Profiles")
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatDefault
	t.AppendHeader(table.Row{"NAME", "DOMAINS", "SLEEP", "JITTER", "GET", "POST", "SLEEP OBF", "HEAP OBF", "SYSCALLS", "STACK SPOOF"})

	for _, p := range profiles {
		var domainStrs []string
		for _, d := range p.Domains {
			scheme := "http"
				if d.IsHttps {
					scheme = "https"
				}
				domainStrs = append(domainStrs, fmt.Sprintf("%s://%s:%d", scheme, d.Domain, d.Port))
		}
		var syscall string
		if p.Syscall == 1 {
			syscall = "direct"
		} else if p.Syscall == 2 {
			syscall = "indirect"
		} else {
			syscall = "none"
		}
		t.AppendRow(table.Row{
			p.Name,
			strings.Join(domainStrs, ","),
			p.Sleep,
			p.Jitter,
			p.Get,
			p.Post,
			p.SleepObf,
			p.HeapObf,
			syscall,
			p.StackSpoof,
		})
	}

	c.ui.Send(t.Render())
}

func (c *CLI) LoadAndSendProfile(args []string) {
	fs := pflag.NewFlagSet("new", pflag.ContinueOnError)
	file := fs.StringP("profile", "p", "", "")
	if err := fs.Parse(args); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error parsing flags: %v", err))
		return
	}

	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".kronos", "profile.json")

	if *file != "" {
		path = *file
	}

	data, err := os.ReadFile(path)
	if err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Getting Profile: %v", err))
		return
	}

	var prof Profile
	if err = json.Unmarshal(data, &prof); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed parsing profile: %s", err))
		return
	}

	if err := c.http.DoPost("ts/rest/agent/profiles/new", data, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed sending profile: %s", err))
		return 
	}

	c.ui.PrintTitle(fmt.Sprint("Profile Saved!"))
	return 

}