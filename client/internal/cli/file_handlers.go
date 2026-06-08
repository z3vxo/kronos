package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/z3vxo/kronos/internal/ui"
)

func (c *CLI) HandleDOWNLOAD(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: download <file>"))
		return
	}

	payload := TaskEntry{
		Guid:     c.ClientInUse,
		Cmd_type: CmdCodeMap["download"],
		Param1:   strings.Join(args, " "),
	}

	data, err := json.Marshal(&payload)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed marshaling json: %s", err))
		return
	}

	if err := c.http.DoPost("ts/rest/file/download/task", data, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed inserting download task: %s", err))
		return
	}

	c.ui.PrintTitle(fmt.Sprintf("Tasked %s", c.ui.InUse))
}

func (c *CLI) HandleFiles(args []string) {
	if len(args) == 0 || args[0] == "list" {
		c.listFiles()
		return
	}
	if args[0] == "sync" {
		if len(args) < 2 {
			c.ui.Send(ui.BAD.Sprint("Usage: files sync <id>"))
			return
		}
		c.syncFile(args[1])
		return
	}
	c.ui.Send(ui.BAD.Sprint("Usage: files [list] | files sync <id>"))
}

func (c *CLI) listFiles() {
	var resp FilesResp
	if err := c.http.DoGet("ts/rest/files/list", &resp); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed listing files: %s", err))
		return
	}
	if resp.Total == 0 {
		c.ui.PrintTitle("No files")
		return
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 6},
		{Number: 2, WidthMin: 16},
		{Number: 3, WidthMin: 30},
		{Number: 4, WidthMin: 10},
	})
	t.AppendHeader(table.Row{"ID", "AGENT", "FILENAME", "SIZE"})
	t.AppendSeparator()
	for _, f := range resp.Files {
		t.AppendRow(table.Row{f.ID, f.AgentID, f.Name, fmtFileSize(f.Size)})
	}
	c.ui.Send(t.Render())
}

func (c *CLI) syncFile(id string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/ts/rest/files/sync/%s", c.http.Hostname, id), nil)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed creating request: %s", err))
		return
	}
	c.http.Auth.Apply(req)

	resp, err := c.http.HttpClient.Do(req)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed downloading file: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.ui.Send(ui.BAD.Sprintf("Server returned: %s", resp.Status))
		return
	}

	filename := id
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if parts := strings.SplitN(cd, "filename=", 2); len(parts) == 2 {
			filename = strings.Trim(parts[1], "\"")
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed resolving home dir: %s", err))
		return
	}
	outDir := filepath.Join(home, ".kronos", "files")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed creating output dir: %s", err))
		return
	}
	outPath := filepath.Join(outDir, filename)

	f, err := os.Create(outPath)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed creating file: %s", err))
		return
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed writing file: %s", err))
		return
	}

	c.ui.PrintTitle(fmt.Sprintf("Saved %s (%s)", outPath, fmtFileSize(uint64(n))))
}

func fmtFileSize(n uint64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f mb", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f kb", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%d b", n)
	}
}
