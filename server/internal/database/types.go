package database

type Agent struct {
	AgentID    uint32    `json:"agent_id"`
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

type Agents struct {
	Total int     `json:"total"`
	Agent []Agent `json:"agents"`
}

type Task struct {
	ID       int    `json:"id"`
	Guid     string `json:"guid"`
	CmdCode  int    `json:"cmd_code"`
	CmdName  string `json:"cmd_name,omitempty"`
	Param1   string `json:"param_1"`
	Param2   string `json:"param_2"`
	Params   []byte `json:"params"`
	TaskID   uint32 `json:"task_id"`
	TaskedAt int    `json:"tasked_at"`
}

type TaskEntrys struct {
	Total int    `json:"total"`
	Tasks []Task `json:"tasks"`
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
	Payload_id  uint32          `json:"pay_id"`
	Key         []byte		    `json:"Key"`
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