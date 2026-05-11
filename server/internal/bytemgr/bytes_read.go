package bytemgr

import (
	"bytes"
	"encoding/binary"
	"io"
	"fmt"
)

type Reader struct {
	r   *bytes.Reader
	err error
}

func (r *Reader) Read4() uint32 {
	if r.err != nil {
		return 0
	}
	var val uint32
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func (r *Reader) Read1() byte {
	if r.err != nil {
		return 0
	}
	val, err := r.r.ReadByte()
	r.err = err
	return val
}

func (r *Reader) Read2() int16 {
	if r.err != nil {
		return 0
	}
	var val int16
	r.err = binary.Read(r.r, binary.LittleEndian, &val)
	return val
}

func (r *Reader) ReadString(len uint32) string {
	if r.err != nil {
		return ""
	}
	buf := make([]byte, len)
	_, r.err = io.ReadFull(r.r, buf)
	return string(buf)

}

/*
	[MSG TYPE]		  4 BYTES
	[HADES ID]		  4 BYTES
	[UserLen]         4 BYTES
	[Username]		  N BYTES
	[HostLen]		  4 BYTES
	[Hostname]		  N BYTES
	[IP LEN]		  4 BYTES
	[IP STR]		  N BYTES
	[ProcessPath Len] 4 BYTES
	[PROCESS PATH]    N BYTES
	[PID]			  4 BYTES
	[TID]			  4 BYTES
	[PPID]			  4 BYTES
	[IsElev]		  1 BYTES
	[Arch]			  1 BYTES
	[Minor]			  4 BYTES
	[Major]			  4 BYTES
	[Build]			  4 BYTES

*/

type ClientRegister struct {
	Guid       uint32
	User       string
	Host       string
	InternaIP  string
	ExternalIP string
	ProcPath   string
	Pid        uint32
	Tid        uint32
	Ppid       uint32
	IsElev     byte
	Arch       byte
	Minor      uint32
	Major      uint32
	Build      uint32
}

func ExtractRegistrationDetails(IP string, r *bytes.Reader) (ClientRegister, error) {
	rd := &Reader{r: r}



	guid := rd.Read4()
	Username := rd.ReadString(rd.Read4())
	Hostname := rd.ReadString(rd.Read4())
	InternalIP := rd.ReadString(rd.Read4())
	ProcessPath := rd.ReadString(rd.Read4())
	Pid := rd.Read4()
	Tid := rd.Read4()
	PPid := rd.Read4()
	IsElev := rd.Read1()
	Arch := rd.Read1()
	Minor := rd.Read4()
	Major := rd.Read4()
	BuildVer := rd.Read4()
	if rd.err != nil {
		fmt.Println(rd.err);
		return ClientRegister{}, rd.err
	}
	fmt.Printf("New Client: %s\n", Username)
	
	Res := ClientRegister{
		Guid:       guid,
		User:       Username,
		Host:       Hostname,
		InternaIP:  InternalIP,
		ExternalIP: IP,
		ProcPath:   ProcessPath,
		Pid:        Pid,
		Tid:        Tid,
		Ppid:       PPid,
		IsElev:     IsElev,
		Arch:       Arch,
		Minor:      Minor,
		Major:      Major,
		Build:      BuildVer,
	}

	return Res, nil
}

type OutputEntrys struct {
	//Type   uint32
	TaskID uint32
	Output []byte
}

/*
  [task output count] 4 bytes
 // looped
 [TASKID] 4 BYTES
 [STATUS] 4 bytes -> if 0 == success read next, if 1 == read read4() for error code
 [TASK_TYPE] 4 BYTES -> parse this, jump to handler and parse it, if task type == 0, continue below, single string output no further parsing
 ------
 [HAS_DATA] 4 BYTES -> if > 1 lookup in success map else below
 [OUTPUT LEN] 4 BYTES
 [OUTPUT DATA] N BYTES
*/

func ParseClientOutput(r *bytes.Reader) ([]OutputEntrys, error) {
	rd := Reader{r: r}
	var entrys []OutputEntrys
	count := rd.Read4()
	for range count {
		fmt.Println("================================")
		var o OutputEntrys
		o.TaskID = rd.Read4()
		status := rd.Read4()
		fmt.Printf("TASK_ID: %d\nStatus: %d\n", o.TaskID, status)

		if status == 1 {

			code := rd.Read4()
			ErrorStr := ErrorCodeMap[code]
 			o.Output = []byte(ErrorStr)
 			entrys = append(entrys, o)
 			continue
		}

		TaskType := rd.Read4()
		fmt.Printf("TASK_TYPE: %d\n", TaskType)
		if TaskType > 0 {
			if TaskType == 3 {
				Total := rd.Read4()
				fmt.Printf("TOTAL: %d\n", Total)
				for range Total {
					EntryLen := rd.Read4()
					entryStr := rd.ReadString(EntryLen)
					typeLen := rd.Read4()
					typeStr := rd.ReadString(typeLen)
					fmt.Printf("Name: %s | type %s\n", entryStr,typeStr)
				}
				continue
			}
		}
		fmt.Println("================================")
		HasData := rd.Read4()
		if HasData > 1 {
			SuccessString := SuccessMap[HasData]
			o.Output = []byte(SuccessString)
			entrys = append(entrys, o)
			continue
		}
		OutputLen := rd.Read4()
		o.Output = []byte(rd.ReadString(OutputLen))
		entrys = append(entrys, o)
	}
		if rd.err != nil {
 		return nil, rd.err
 	}

 	return entrys, nil
}




/*
	[total] 4 bytes
	repeated
	[entry len] 4 bytes
	[entry string] N bytes
	[type len] 4 bytes
	[type string] N bytes
*/

func ParseLSOutput(r *Reader) string {
	fmt.Println("Parsing LS")

	Total := r.Read4()
	fmt.Println(Total)
	for range Total {
		EntryLen := r.Read4()
		entryStr := r.ReadString(EntryLen)
		typeLen := r.Read4()
		typeStr := r.ReadString(typeLen)
		fmt.Printf("Name: %s | type %s\n", entryStr,typeStr)
	}
	return ""
}
