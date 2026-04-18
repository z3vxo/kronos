package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

func (c *CLI) ListAgents(args []string) {
	var A Agents
	if err := c.http.DoGet("ts/rest/agents/list", &A); err != nil {
		fmt.Println("[!] Failed listing agents:", err)
		return
	}

	if A.Total == 0 {
		fmt.Println("[*] No agents connected")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Print("\r\033[K")
	fmt.Fprintln(w, "\nCODENAME\tUSER\tHOSTNAME\tEXT IP\tINT IP\tELEV\tPID\tLAST SEEN")
	fmt.Fprintln(w, "--------\t----\t--------\t------\t------\t----\t---\t---------")
	for _, a := range A.Agent {
		elev := "no"
		if a.IsElevated {
			elev = "yes"
		}
		last := time.Unix(a.LastSeen, 0).Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			a.CodeName, a.Username, a.Hostname, a.Ex_ip, a.In_ip, elev, a.Pid, last)
	}

	w.Flush()
}
