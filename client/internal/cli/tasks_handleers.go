package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/z3vxo/kronos/internal/ui"
)



func (c *CLI) ParseTasks(args []string) {
	if len(args) == 0 {
		c.ListTasks()
		return
	}
	if args[0] == "delete" {
		if len(args) < 2 || args[1] == "" {
			c.ui.Send(ui.WARN.Sprint("Must Provide ID or TaskID"))
			return
		}
		c.DeleteTask(args[1])
		return
	}

}

func (c *CLI) ListTasks() {
	if c.ClientInUse == "" {
		c.ui.Send(ui.WARN.Sprint("Must be using agent"))
		return
	}
	var Task TaskEntrys
	if err := c.http.DoGet(fmt.Sprintf("ts/rest/tasks/list/%s", c.ClientInUse), &Task); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed Listing Tasks: %s", err))
		return
	}
	if len(Task.Tasks) == 0 {
		c.ui.PrintTitle("No Tasks")
		return
	}

	c.CacheMgr.TaskIdMap = make(map[uint32]uint32)

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 6},
		{Number: 2, WidthMin: 8},
		{Number: 3, WidthMin: 20},
		{Number: 4, WidthMin: 20},
		{Number: 5, WidthMin: 12},
		{Number: 6, WidthMin: 14},
	})
	t.AppendHeader(table.Row{"ID", "CMD", "ARGS", "TASK-ID", "TASKED-AT"})
	t.AppendSeparator()
	for _, i := range Task.Tasks {
		c.CacheMgr.TaskIdMap[i.ID] = i.TaskID
		args := i.Param1 + " " + i.Param2
		t.AppendRow(table.Row{
			i.ID,
			i.CmdName,
			args,
			i.TaskID,
			relativeTime(int64(i.TaskedAt)),
		})
	}
	c.ui.Send(t.Render())

}

func (c *CLI) DeleteTask(id string) {
	var name uint32
	ID, err := strconv.ParseUint(id, 10, 32)
	if err == nil {
		taskId, ok := c.CacheMgr.TaskIdMap[uint32(ID)]
		if !ok {
			c.ui.Send(ui.BAD.Sprint("Unknown agent ID, run 'tasks' to view or refresh"))
			return
		}
		name = taskId

	} else {
		name = uint32(ID)
	}
	

	if err := c.http.DoDelete(fmt.Sprintf("ts/rest/tasks/delete/%s/%d", c.ClientInUse, name), nil); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error Deleting Task: %s", err))
		return
	}

	c.ui.PrintTitle("Deleted Task!")

}

func (c *CLI) HandleHistory(args []string) {
	if len(args) == 0 {
		c.GetHistory()
		return
	}

	if args[0] == "output" {
		if len(args) < 2 || args[1] == "" {
			c.ui.Send(ui.WARN.Sprint("Must Provide TaskID"))
			return
		}
		c.GetHistoryOutput(args[1])
		return
	}

	c.ui.Send(ui.BAD.Sprint("Unknown sub command"))
}

func (c *CLI) GetHistoryOutput(id string) {
	if !c.requireAgent() {
		return
	}

	resp, err := c.http.DoGetRaw(fmt.Sprintf("ts/rest/tasks/history/%s/%s", c.ClientInUse, id))
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error getting output: %s", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error reading output: %s", err))
		return
	}

	c.ui.PrintTitle(fmt.Sprintf("Output for TaskID: %s\n", id))
	c.ui.Send(string(body))
}


func (c *CLI) GetHistory() {
	if !c.requireAgent() {
		return
	}
	var h History

	if err := c.http.DoGet(fmt.Sprintf("ts/rest/tasks/history/%s", c.ClientInUse), &h); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error listing history: %s", err))
		return
	}

	if h.Total == 0 {
		c.ui.Send(ui.INFO.Sprint("No history for agent"))
		return
	}

	c.ui.PrintTitle(fmt.Sprintf("Task History (%d)", h.Total))
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"#", "TASK-ID", "CMD", "ARGS", "TASKED", "FINISHED"})
	t.AppendSeparator()
	for i, e := range h.Entry {
		args := e.Param1 + " " + e.Param2
		t.AppendRow(table.Row{
			i + 1,
			e.TaskID,
			e.CmdName,
			args,
			relativeTime(e.TaskedAt),
			relativeTime(e.FinishedAt),
		})
	}
	c.ui.Send(t.Render())
}