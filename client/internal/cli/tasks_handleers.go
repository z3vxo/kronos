package cli

import (
	"fmt"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/z3vxo/kronos/internal/ui"
)

var IdToCmdMap = map[int]string{
	0: "ps",
	1: "cmd",
	2: "cat",
	3: "ls",
	4: "rm",
	5: "mv",
}

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

	c.CacheMgr.TaskIdMap = make(map[int]string)

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
	t.AppendHeader(table.Row{"ID", "CMD", "PARAM1", "PARAM2", "TASK-ID", "TASKED-AT"})
	t.AppendSeparator()
	for _, i := range Task.Tasks {
		c.CacheMgr.TaskIdMap[i.ID] = i.TaskID
		if i.Param1 == "" {
			i.Param1 = "NULL"
		}
		if i.Param2 == "" {
			i.Param2 = "NULL"
		}
		name := IdToCmdMap[i.CmdCode]
		t.AppendRow(table.Row{
			i.ID,
			name,
			i.Param1,
			i.Param2,
			i.TaskID,
			relativeTime(int64(i.TaskedAt)),
		})
	}
	c.ui.Send(t.Render())

}

func (c *CLI) DeleteTask(id string) {
	var name string
	if ID, err := strconv.Atoi(id); err == nil {
		taskId, ok := c.CacheMgr.TaskIdMap[ID]
		if !ok {
			c.ui.Send(ui.BAD.Sprint("Unknown agent ID, run 'tasks' to view or refresh"))
			return
		}
		name = taskId

	} else {
		name = id
	}

	if err := c.http.DoDelete(fmt.Sprintf("ts/rest/tasks/delete/%s/%s", c.ClientInUse, name), nil); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Error Deleting Task: %s", err))
		return
	}

	c.ui.PrintTitle("Deleted Task!")

}
