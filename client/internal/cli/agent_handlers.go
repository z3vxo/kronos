package cli

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/z3vxo/kronos/internal/ui"
)

const (
	dim = "\001\033[2m\033[4m\002"
	rst = "\001\033[0m\002"
)

func (c *CLI) PrintTitle(msg string) {
	str := fmt.Sprintf("%s[%s]%s \033[1;36m[*]\033[0m %s", dim, time.Now().Format("2/1 15:04:05"), rst, msg)
	c.ui.Send(str)
}

func relativeTime(unix int64) string {
	since := time.Since(time.Unix(unix, 0))
	switch {
	case since < time.Minute:
		return fmt.Sprintf("%ds", int(since.Seconds()))
	case since < time.Hour:
		return fmt.Sprintf("%dm", int(since.Minutes()))
	case since < 24*time.Hour:
		return fmt.Sprintf("%dh", int(since.Hours()))
	default:
		return fmt.Sprintf("%dd", int(since.Hours()/24))
	}
}

func (c *CLI) ListAgents(args []string) {
	var A Agents
	if err := c.http.DoGet("ts/rest/agents/list", &A); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed listing agents: %v", err))
		return
	}

	if A.Total == 0 {
		c.ui.Send(ui.INFO.Sprint("No agents connected"))
		return
	}
	c.PrintTitle("Active Agents")
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"CODENAME", "USER", "HOSTNAME", "EX IP", "IN IP", "ELEV", "PID", "LAST-SEEN", "REG-DATE"})

	for _, a := range A.Agent {
		elev := "no"
		if a.IsElevated {
			elev = "yes"
		}
		last := relativeTime(a.LastSeen)
		reg := time.Unix(a.RegDate, 0).Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			a.CodeName,
			a.Username,
			a.Hostname,
			a.Ex_ip,
			a.In_ip,
			elev,
			a.Pid,
			last,
			reg,
		})
	}

	c.ui.Send(t.Render())
}

func (c *CLI) ListAgentInfo(args []string) {
	if c.ui.InUse == "" {
		c.ui.Send(ui.BAD.Sprint("Must be using agent!"))
		return
	}
	var a AgentInfoResp

	if err := c.http.DoGet(fmt.Sprintf("ts/rest/agents/info/%s", c.ui.InUse), &a); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed listing info: %v", err))
		return
	}
	c.PrintTitle(fmt.Sprintf("Info for %s", c.ui.InUse))

	elev := "no"
	if a.IsElevated {
		elev = "yes"
	}
	arch := "x86"
	if a.Arch == 1 {
		arch = "x64"
	}
	last := relativeTime(a.LastCheckin)
	reg := time.Unix(a.RegisterTime, 0).Format("2006-01-02 15:04:05")

	t := table.NewWriter()

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 18},
		{Number: 2, WidthMin: 40},
	})

	t.AppendRow(table.Row{"AGENT", ""}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignCenter})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"CODENAME", c.ui.InUse},
		{"USER", a.User},
		{"HOSTNAME", a.Host},
		{"INT IP", a.InternalIP},
		{"EXT IP", a.ExternalIP},
		{"PID / PPID", fmt.Sprintf("%d / %d", a.Pid, a.PPid)},
		{"ARCH", arch},
		{"ELEVATED", elev},
		{"OS", a.WinVer},
		{"PROCESS", a.ProcPath},
		{"LAST SEEN", last},
		{"REGISTERED", reg},
	})

	t.AppendSeparator()
	t.AppendRow(table.Row{"EVASION", ""}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignCenter})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"Heap Obfuscation", "true"},
		{"Sleep", "15"},
		{"Jitter", "10%"},
	})

	t.SetStyle(table.StyleLight)
	c.ui.Send(t.Render())
}

func (c *CLI) Back(args []string) {
	c.ui.Send(ui.INFO.Sprintf("Not using %s", c.ClientInUse))
	c.ClientInUse = ""
	c.ui.InUse = ""
	c.ui.SetPrompt("")
}

func (c *CLI) ResolveAgent(args []string) {
	if len(args) < 1 || args[0] == "" {
		c.ui.Send(ui.BAD.Sprint("Error: must choose agent"))
		return
	}

	var r ResolveResp
	e := fmt.Sprintf("ts/rest/agents/resolve/%s", args[0])

	if err := c.http.DoGet(e, &r); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed resolving agent: %v", err))
		return
	}

	if r.Guid == "" {
		c.ui.Send(ui.BAD.Sprint("Server Did not return a guid!"))
		return
	}

	c.ClientInUse = r.Guid
	c.ui.InUse = args[0]
	c.ui.SetPrompt(args[0])
	c.ui.Send(ui.GOOD.Sprintf("Using %s", c.ClientInUse))
	return
}
