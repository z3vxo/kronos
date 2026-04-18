package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/z3vxo/kronos/internal/httpclient"
)

const prompt = "\001\033[2m\033[4m\002kronos\001\033[0m\002 $> "

type CLI struct {
	http          *httpclient.Client
	rl            *readline.Instance
	Out           *Output
	ClientInUse   string
	dispatchTable map[string]HandlerFunc
}

type HandlerFunc func(args []string)

func NewCli() (*CLI, error) {
	h, err := httpclient.NewClient()
	if err != nil {
		return nil, err
	}

	rl, err := readline.New(prompt)
	if err != nil {
		return nil, err
	}

	c := &CLI{
		http: h,
		rl:   rl,
		Out:  &Output{ch: make(chan string, 32), prompt: prompt},
	}
	go c.Out.Run()
	c.SetupDispatchTable()

	go h.ConnectToSSE()

	return c, nil
}

func (c *CLI) SetupDispatchTable() {
	c.dispatchTable = map[string]HandlerFunc{
		"list": c.ListAgents,
	}
}

func (c *CLI) Close() {
	c.rl.Close()
}

func (c *CLI) Split(input string) ([]string, error) {
	return shlex.Split(input)
}

func (c *CLI) Dispatch(cmd []string) {
	fn, ok := c.dispatchTable[cmd[0]]
	if !ok {
		c.Out.Send(fmt.Sprintf("[!] Unknown command: %s", cmd[0]))
		return
	}
	go fn(cmd[1:])
}

func (c *CLI) Run() {
	defer c.Close()
	for {
		input, err := c.rl.Readline()
		if err == io.EOF || err == readline.ErrInterrupt {
			break
		}
		if err != nil {
			c.Out.Send(fmt.Sprintf("error: %v", err))
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		cmd, err := c.Split(input)
		if err != nil {
			c.Out.Send("[!] Failed parsing input")
			continue
		}

		if strings.ToLower(cmd[0]) == "exit" {
			c.Close()
			os.Exit(0)
		}

		c.rl.SaveHistory(input)
		c.Dispatch(cmd)
	}
}
