package cli

// ---- AGENT STRUCTS ----

type Agent struct {
	AgentID    uint32  `json:"agent_id"`
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
	Pid          uint32  `json:"pid"`
	PPid         uint32  `json:"ppid"`
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

type Param struct {
	Value    any    `json:"param"`
	DataType string `json:"type"`
}

// ---- TASK STRUCTS ----
type TaskEntry struct {
	Cmd_type int     `json:"type"`
	Guid     string  `json:"guid"`
	Params   []Param `json:"params"`
}

type Task struct {
	ID       uint32 `json:"id"`
	Guid     string `json:"guid"`
	CmdCode  int    `json:"cmd_code"`
	CmdName  string `json:"cmd_name"`
	Param1   string `json:"param_1"`
	Param2   string `json:"param_2"`
	TaskID   uint32 `json:"task_id"`
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

// ----- Files -----

type FileEntry struct {
	ID      uint32 `json:"id"`
	AgentID string `json:"agent_id"`
	Name    string `json:"filename"`
	Size    uint64 `json:"size"`
}

type FilesResp struct {
	Total int         `json:"total"`
	Files []FileEntry `json:"files"`
}

type UploadStartReq struct {
	AgentID  string `json:"agentid"`
	Path     string `json:"path"`
	FileSize int64  `json:"size"`
}

type UploadStartResp struct {
	UploadID string `json:"upload_id"`
}


type ProfileDomain struct {
	Domain  string `json:"domain_value"`
	Port    int    `json:"port"`
	IsHttps bool   `json:"is_https"`
}

type ProfileHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Profile struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Domains     []ProfileDomain `json:"domains"`
	Headers     []ProfileHeader `json:"headers"`
	Sleep       int             `json:"sleep"`
	Jitter      int             `json:"jitter"`
	Get         string          `json:"get_endpoint"`
	Post        string          `json:"post_endpoint"`
	SleepObf    bool            `json:"sleep_obf"`
	HeapObf     bool            `json:"heap_obf"`
	StackSpoof  bool            `json:"stack_spoof"`
	Syscall     int             `json:"syscall"`
}

type GeneratePayloadReq struct {
	Name string `json:"name"`
	Debug bool `json:"debug"`
	Format string `json:"format"`
}


type HistoryEntry struct {
	CmdCode    int    `json:"cmd_code"`
	CmdName    string `json:"cmd_name"`
	TaskID     uint32 `json:"task_id"`
	Param1     string `json:"param_1"`
	Param2     string `json:"param_2"`
	TaskedAt   int64  `json:"tasked_at"`
	FinishedAt int64  `json:"finished_at"`
}

type History struct {
	Total int            `json:"total"`
	Entry []HistoryEntry `json:"history"`
}