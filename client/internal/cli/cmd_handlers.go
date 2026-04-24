package cli

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/z3vxo/kronos/internal/ui"
)

var CmdCodeMap = map[string]int{
	"ps":  0,
	"cmd": 1,
	"cat": 2,
	"ls":  3,
	"rm":  4,
	"mv":  5,
}

func genTaskID() int {
	return rand.Intn(90000000) + 10000000
}

func (c *CLI) requireAgent() bool {
	if c.ClientInUse == "" {
		c.ui.Send(ui.BAD.Sprint("Must be using agent"))
		return false
	}
	return true
}

func (c *CLI) HandlePS(args []string) {
	if !c.requireAgent() {
		return
	}
	if len(args) == 0 {
		c.ui.Send(ui.BAD.Sprint("Usage: ps <args>"))
		return
	}

	cmdCode := CmdCodeMap["ps"]
	param1 := strings.Join(args, " ")
	taskID := genTaskID()
	payload := TaskEntry{
		Guid:     c.ClientInUse,
		TaskID:   taskID,
		Cmd_type: cmdCode,
		Param1:   param1,
	}

	data, err := json.Marshal(&payload)
	if err != nil {
		c.ui.Send(ui.BAD.Sprint("Failed marshaling json"))
		return
	}

	if err := c.http.DoPost("ts/rest/tasks/new", data, nil); err != nil {
		c.ui.Send(ui.WARN.Sprintf("Failed Inserting command: %s", err))
	}

	c.ui.PrintTitle(fmt.Sprintf("Tasked %s", c.ui.InUse))

}
