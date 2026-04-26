package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/google/shlex"
	"github.com/z3vxo/kronos/internal/ai"
	"github.com/z3vxo/kronos/internal/httpclient"
	"github.com/z3vxo/kronos/internal/ui"
)

type CLI struct {
	http          *httpclient.Client
	ui            *ui.UI
	ai            *ai.AI
	ClientInUse   string
	dispatchTable map[string]HandlerFunc
	CacheMgr      Cache
}

type HandlerFunc func(args []string)

func NewCli() (*CLI, error) {
	rl, err := ui.NewUI()
	if err != nil {
		return nil, err
	}
	h, err := httpclient.NewClient(rl)
	if err != nil {
		return nil, err
	}

	ai, _ := ai.NewAI(rl)
	c := &CLI{
		http: h,
		ui:   rl,
		ai:   ai,
	}
	h.InvalidateAgentCache = func() { c.CacheMgr.InvalidateAgents() }
	go c.ui.Run()
	c.SetupDispatchTable()

	go h.ConnectToSSE()

	return c, nil
}

func (c *CLI) SetupDispatchTable() {
	c.dispatchTable = map[string]HandlerFunc{
		"agents":    c.ListAgents,
		"use":       c.ResolveAgent,
		"back":      c.Back,
		"listeners": c.ParseListenerCmd,
		"info":      c.ListAgentInfo,
		"help":      c.Help,
		"ps":        c.HandlePS,
		"ai":        c.ai.HandleAI,
		"tasks":     c.ParseTasks,
	}
}

func (c *CLI) Close() {
	c.ui.Rl.Close()
}

func (c *CLI) Dispatch(cmd []string) {
	fn, ok := c.dispatchTable[cmd[0]]
	if !ok {
		c.ui.Send(fmt.Sprintf("[!] Unknown command: %s", cmd[0]))
		return
	}
	go fn(cmd[1:])
}

func (c *CLI) Run() {
	defer c.Close()
	for {
		input, err := c.ui.Rl.Readline()
		if err == io.EOF || err == readline.ErrInterrupt {
			break
		}
		if err != nil {
			c.ui.Send(fmt.Sprintf("error: %v", err))
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		cmd, err := shlex.Split(input)
		if err != nil {
			c.ui.Send(ui.BAD.Sprint("Failed parsing input"))
			continue
		}

		if strings.ToLower(cmd[0]) == "exit" {
			c.Close()
			os.Exit(0)
		}
		c.ui.Rl.SaveHistory(input)

		c.Dispatch(cmd)
	}
}

func (c *CLI) Help(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "ai":
			c.helpAI()
		case "listeners":
			c.helpListeners()
		case "agents":
			c.helpAgents()
		case "commands":
			c.helpCommands()
		default:
			c.ui.Send(fmt.Sprintf("[!] No help for '%s'. Topics: agents, listeners, commands, ai", args[0]))
		}
		return
	}

	c.ui.Send("\n")
	c.ui.Send("\033[1;35m  ██╗  ██╗██████╗  ██████╗ ███╗   ██╗ ██████╗ ███████╗\033[0m")
	c.ui.Send("\033[1;35m  ██║ ██╔╝██╔══██╗██╔═══██╗████╗  ██║██╔═══██╗██╔════╝\033[0m")
	c.ui.Send("\033[1;35m  █████╔╝ ██████╔╝██║   ██║██╔██╗ ██║██║   ██║███████╗\033[0m")
	c.ui.Send("\033[1;35m  ██╔═██╗ ██╔══██╗██║   ██║██║╚██╗██║██║   ██║╚════██║\033[0m")
	c.ui.Send("\033[1;35m  ██║  ██╗██║  ██║╚██████╔╝██║ ╚████║╚██████╔╝███████║\033[0m")
	c.ui.Send("\033[1;35m  ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝\033[0m")
	c.ui.Send("")
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-20s\033[0m %s", "help agents", "agent interaction commands"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-20s\033[0m %s", "help listeners", "listener management commands"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-20s\033[0m %s", "help commands", "agent commands (shell, fs, etc.)"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-20s\033[0m %s", "help ai", "AI command generation"))
	c.ui.Send("")
}

func (c *CLI) helpAgents() {
	c.ui.Send("\033[1;37m  AGENTS\033[0m")
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "list", "list all connected agents"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "use <codename>", "interact with an agent"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "info", "detailed info on current agent"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "back", "stop using current agent"))
}

func (c *CLI) helpListeners() {
	c.ui.Send("\033[1;37m  LISTENERS\033[0m")
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "listeners", "list all listeners"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "listeners new -h <host> -p <port> -t <proto>", "create listener (proto: http|https)"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "listeners start <name>", "start an inactive listener"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "listeners stop <name>", "stop a running listener"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "listeners delete <name>", "delete a listener"))
}

func (c *CLI) helpCommands() {
	c.ui.Send("\033[1;37m  COMMANDS\033[0m")
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "whoami", "current user identity"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "getprivs", "current user privileges"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ls <dir>", "list a directory"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "cd <dir>", "change directory"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "cat <file>", "read a file"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ps", "list running processes"))
}

func (c *CLI) helpAI() {
	c.ui.Send("\033[1;37m  AI\033[0m")
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ai <prompt>", "generate a command from natural language"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ai models", "list available models and pricing"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ai settings", "show current provider/model/temp"))
	c.ui.Send(fmt.Sprintf("  \033[1;36m%-40s\033[0m %s", "ai reconfig -p <provider> -m <model> [-t <temp>]", "reconfigure AI provider"))
}
