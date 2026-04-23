package cli

import (
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/pflag"
	"github.com/z3vxo/kronos/internal/ui"
)

func (c *CLI) ParseListenerCmd(args []string) {
	if len(args) == 0 {
		c.ListListners()
		return
	}

	switch args[0] {
	case "stop":
		if args[1] == "" {
			c.ui.Send(ui.BAD.Sprint("Must provide listener name"))
		}
		c.StopListener(args[1])
	case "delete":
		if args[1] == "" {
			c.ui.Send(ui.BAD.Sprint("Must provide listener name"))
		}
		c.DeleteListeners(args[1])
	case "start":
		if args[1] == "" {
			c.ui.Send(ui.BAD.Sprint("Must provide listener name"))
		}
		c.StartListener(args[1])

	case "new":
		c.NewListener(args[1:])
		//c.StopListener()
	default:
		c.ui.Send(ui.BAD.Sprintf("Unknown subcommand: %s", args[0]))
	}
}

func (c *CLI) DeleteListeners(name string) {
	endpoint := fmt.Sprintf("ts/rest/listeners/delete/%s", name)
	if err := c.http.DoDelete(endpoint, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Deleting listener: %s", err))
		return
	}
	c.ui.Send(ui.GOOD.Sprintf("Deleted Listener %s", name))
	return

}

func (c *CLI) StopListener(name string) {
	endpoint := fmt.Sprintf("ts/rest/listeners/stop/%s", name)
	if err := c.http.DoPost(endpoint, nil, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Stopping Listener: %s", err))
		return
	}

	c.ui.Send(ui.GOOD.Sprintf("Stopped Listener %s", name))
	return

}

func (c *CLI) ListListners() {
	var r ListListenersResp

	if err := c.http.DoGet("ts/rest/listeners/list", &r); err != nil {
		c.ui.Send(ui.BAD.Sprint("Failed Listing Listeners!"))
		return
	}

	if r.Total == 0 {
		c.ui.Send(ui.INFO.Sprint("No Active Listeners"))
		return
	}

	c.PrintTitle("Active Listeners")

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{
		"NAME", "PORT", "PROTOCOL", "HOST", "STATUS",
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 20},
		{Number: 2, WidthMin: 8, Align: text.AlignLeft, AlignHeader: text.AlignLeft},
		{Number: 3, WidthMin: 12},
		{Number: 4, WidthMin: 16},
		{Number: 5, WidthMin: 12},
	})

	for _, i := range r.Listeners {
		status := "RUNNING"
		if i.Status == false {
			status = "STOPPED"
		}
		t.AppendRow(table.Row{
			i.Name,
			i.Port,
			i.Protocol,
			i.Host,
			status,
		})
	}

	c.ui.Send(t.Render())
}

func (c *CLI) StartListener(name string) {
	endpoint := fmt.Sprintf("ts/rest/listeners/start/%s", name)
	if err := c.http.DoPost(endpoint, nil, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Deleting listener: %s", err))
		return
	}
	c.ui.Send(ui.GOOD.Sprintf("Started Listener %s", name))
	return
}

func (c *CLI) NewListener(args []string) {
	fs := pflag.NewFlagSet("start", pflag.ContinueOnError)
	port := fs.IntP("port", "p", 443, "")
	proto := fs.StringP("type", "t", "http", "")
	host := fs.StringP("host", "h", "", "")
	cert := fs.BoolP("lets-encrypt", "l", false, "")
	if err := fs.Parse(args); err != nil {
		c.ui.Send(ui.WARN.Sprintf("[!] %v", err))
		return
	}

	if *host == "" {
		c.ui.Send(ui.WARN.Sprint("Must provide host"))
		return
	}

	data := ListenStartReq{
		Port:     *port,
		Protocol: *proto,
		Host:     *host,
		CertType: *cert,
	}

	body, err := json.Marshal(data)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error Marshaling json: %v", err))
		return
	}
	var StartResp ListenerStartResp
	if err := c.http.DoPost("ts/rest/listeners/new", body, &StartResp); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error Starting Listener: %v", err))
		return

	}

	c.ui.Send(ui.GOOD.Sprintf("Listener Started: %s", StartResp.Name))
	return
}
