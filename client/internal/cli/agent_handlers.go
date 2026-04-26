package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/z3vxo/kronos/internal/ui"
)

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
	if cached := c.CacheMgr.GetAgentsCache(); cached != nil {
		fmt.Println("Printing from cache")
		c.PrintAgents(cached)
		return
	}

	var A Agents
	if err := c.http.DoGet("ts/rest/agents/list", &A); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed listing agents: %v", err))
		return
	}

	if A.Total == 0 {
		c.ui.Send(ui.INFO.Sprint("No agents connected"))
		return
	}
	fmt.Println("Not Printing from cache")
	c.CacheMgr.PopulateAgentsCache(A)
	c.PrintAgents(c.CacheMgr.GetAgentsCache())
}

func (c *CLI) PrintAgents(agents []Agent) {
	if len(agents) == 0 {
		c.ui.PrintTitle("No Agents Connected")

		return
	}
	c.ui.PrintTitle("Active Agents")
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"ID", "CODENAME", "USER", "HOSTNAME", "EX IP", "IN IP", "ELEV", "PID", "LAST-SEEN", "REG-DATE"})

	for _, a := range agents {
		elev := "no"
		if a.IsElevated {
			elev = "yes"
		}
		last := relativeTime(a.LastSeen)
		reg := time.Unix(a.RegDate, 0).Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			a.AgentID,
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

	if cached := c.CacheMgr.GetInfoCache(); cached != nil {
		fmt.Println("Printing from cache")
		c.PrintAgentInfo(*cached)
		return
	}

	var a AgentInfoResp
	if err := c.http.DoGet(fmt.Sprintf("ts/rest/agents/info/%s", c.ui.InUse), &a); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed listing info: %v", err))
		return
	}

	c.CacheMgr.PopulateInfoCache(a)
	fmt.Println("not Printing from cache")
	c.PrintAgentInfo(*c.CacheMgr.GetInfoCache())
}

func (c *CLI) PrintAgentInfo(a InfoCache) {
	c.ui.PrintTitle(fmt.Sprintf("Info for %s", c.ui.InUse))

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
	t.SetStyle(table.StyleLight)
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

	c.ui.Send(t.Render())
}

func (c *CLI) Back(args []string) {
	c.ui.Send(ui.INFO.Sprintf("Not using %s", c.ClientInUse))
	c.ClientInUse = ""
	c.ui.InUse = ""
	c.CacheMgr.InvalidateInfo()
	c.ui.SetPrompt("")
}

func (c *CLI) ResolveAgent(args []string) {
	if len(args) < 1 || args[0] == "" {
		c.ui.Send(ui.BAD.Sprint("Error: must choose agent"))
		return
	}

	var name string
	if id, err := strconv.Atoi(args[0]); err == nil {
		codename, ok := c.CacheMgr.ResolveAgentID(id)
		if !ok {
			c.ui.Send(ui.BAD.Sprint("Unknown agent ID, run 'list' to refresh"))
			return
		}
		name = codename
	} else {
		name = args[0]
	}

	var r ResolveResp
	e := fmt.Sprintf("ts/rest/agents/resolve/%s", name)

	if err := c.http.DoGet(e, &r); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed resolving agent: %v", err))
		return
	}

	if r.Guid == "" {
		c.ui.Send(ui.BAD.Sprint("Server Did not return a guid!"))
		return
	}

	c.ClientInUse = r.Guid
	c.ui.InUse = name
	c.ui.SetPrompt(name)
	c.ui.Send(ui.GOOD.Sprintf("Using %s", c.ClientInUse))
}
