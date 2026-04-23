package httpclient

import (
	"net/http"

	"github.com/z3vxo/kronos/internal/auth"
	"github.com/z3vxo/kronos/internal/ui"
)

type ErrorRes struct {
	ErrorStr string `json:"error"`
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

const (
	TYPE_NEW_AGENT  = 1
	TYPE_CMD_OUTPUT = 2
	TYPE_HEARTBEAT  = 3
)

type Client struct {
	Hostname   string
	Auth       *auth.AUTH
	HttpClient *http.Client
	Stream     *http.Client
	UI         *ui.UI
}
