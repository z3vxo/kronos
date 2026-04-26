package cli

// ---- AGENT STRUCTS ----

type Agent struct {
	AgentID    int32  `json:"agent_id"`
	CodeName   string `json:"code_name"`
	Username   string `json:"username"`
	Hostname   string `json:"hostname"`
	Ex_ip      string `json:"ex_ip"`
	In_ip      string `json:"in_ip"`
	IsElevated bool   `json:"is_elevated"`
	Pid        int    `json:"pid"`
	ProcPath   string `json:"proc_path"`
	WinVer     string `json:"winver"`
	LastSeen   int64  `json:"last_checkin"`
	RegDate    int64  `json:"reg_date"`
}

type AgentInfoResp struct {
	User         string `json:"username"`
	Host         string `json:"hostname"`
	ProcPath     string `json:"proc_path"`
	Pid          int32  `json:"pid"`
	PPid         int32  `json:"ppid"`
	WinVer       string `json:"win_version"`
	InternalIP   string `json:"internal_ip"`
	ExternalIP   string `json:"external_ip"`
	IsElevated   bool   `json:"is_elev"`
	Arch         byte   `json:"arch"`
	LastCheckin  int64  `json:"last_checkin"`
	RegisterTime int64  `json:"reg_date"`
}

type Agents struct {
	Total int     `json:"total"`
	Agent []Agent `json:"agents"`
}

type ResolveResp struct {
	Guid string `json:"guid"`
}

// ---- TASK STRUCTS ----
type TaskEntry struct {
	Cmd_type int    `json:"type"`
	Guid     string `json:"guid"`
	TaskID   int    `json:"task_id"`
	Param1   string `json:"param_1"`
	Param2   string `json:"param_2"`
}

type Task struct {
	ID       int    `json:"id"`
	Guid     string `json:"guid"`
	CmdCode  int    `json:"cmd_code"`
	Param1   string `json:"param_1"`
	Param2   string `json:"param_2"`
	TaskID   string `json:"task_id"`
	TaskedAt int    `json:"tasked_at"`
}

type TaskEntrys struct {
	Total int    `json:"total"`
	Tasks []Task `json:"tasks"`
}

// ----- Listener List response data -----
type ListenerEntry struct {
	ID       int
	Port     int
	Name     string
	Protocol string
	Status   bool
	Host     string
}

type ListListenersResp struct {
	Total     int             `json:"total"`
	Listeners []ListenerEntry `json:"listeners"`
}

//----- listener Start request data -----

type ListenStartReq struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	CertType bool   `json:"letsencrypt"` // 0 = self signed, 1 = lets encrypt
}

// Listener Start Response data
type ListenerStartResp struct {
	Name string `json:"listener_name"`
}

type Generic200 struct {
	Status string `json:"status"`
}
