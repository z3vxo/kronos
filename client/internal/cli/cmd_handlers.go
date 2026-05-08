package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/z3vxo/kronos/internal/ui"
)

var CmdCodeMap = map[string]int{
	"ps":  	 0,
	"cmd": 	 1,
	"cat": 	 2,
	"ls":  	 3,
	"rm":  	 4,
	"mv":  	 5,
	"pwd": 	 6,
	"cd":  	 7,
	"cp": 	 8,
	"rmdir": 9,
	"get-privs": 10,
}


func (c *CLI) requireAgent() bool {
	if c.ClientInUse == "" {
		c.ui.Send(ui.BAD.Sprint("Must be using agent"))
		return false
	}
	return true
}

func (c *CLI) sendTask(cmd string, param1 string, param2 string) bool {
	payload := TaskEntry{
		Guid:     c.ClientInUse,
		Cmd_type: CmdCodeMap[cmd],
		Param1:   param1,
		Param2:   param2,
	}

	data, err := json.Marshal(&payload)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed marshaling json: %s", err))
		return false
	}

	if err := c.http.DoPost("ts/rest/tasks/new", data, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Inserting command: %s", err))
		return false
	}

	c.ui.PrintTitle(fmt.Sprintf("Tasked %s", c.ui.InUse))
	return true
}

func(c *CLI) HandleRMDIR(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rmdir <dir>"))
		return
	}
	c.sendTask("rmdir", strings.Join(args, " "), "")
}

func(c *CLI) HandleGETPRIVS(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("get-privs", " ", "")
}

func(c *CLI) HandleRM(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rm <file>"))
		return
	}
	c.sendTask("rm", strings.Join(args, " "), "")
}


func(c *CLI) HandleCAT(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: cat <file>"))
		return
	}
	c.sendTask("cat", strings.Join(args, " "), "")
}

func (c *CLI) HandlePS(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: ps <args>"))
		return
	}

	c.sendTask("ps", strings.Join(args, " "), "")
}

func (c *CLI) HandlePWD(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("pwd", "", "")
}

func (c *CLI) HandleCD(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 1 {
		c.ui.Send(ui.BAD.Sprint("Usage: cd <dir>"))
		return
	}

	c.sendTask("cd", args[0], "")
}

func (c *CLI) HandleCP(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: cp <old> <new>"))
		return
	}

	c.sendTask("cp", args[0], args[1])
}


func (c *CLI) HandleMV(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: mv <src> <dst>"))
		return
	}

	c.sendTask("mv", args[0], args[1])
}
