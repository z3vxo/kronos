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

// Looped
/* [OUTPUT COUNT] 4 BYTES
 * ->
 	[TASK ID] 4 bytes
 	->
 	[Status] 4 bytes
 		- if 1 -> read4() to get error code
 		-if 0 below
 	[HAS DATA] 4 BYTES
 		- if 1 -> done 
 		- if 0 -> read4() to get len -> readString to get data
 */

func ParseClientOutput(r *bytes.Reader) ([]OutputEntrys, error) {

	rd := Reader{r: r}
	var Entrys []OutputEntrys
	Count := rd.Read4()

	for range Count {
		var o OutputEntrys
		o.TaskID = rd.Read4()
		Status := rd.Read4()
		if Status == 1 {
			errorCode := rd.Read4()
			ErrorStr := ErrorCodeMap[errorCode]
			o.Output = []byte(ErrorStr)
			Entrys = append(Entrys, o)
			continue
		}
		hasData := rd.Read4()
		if hasData == 1 {
			o.Output = []byte("Agent Completed task")
			Entrys = append(Entrys, o)
			continue
		}
		dataLen := rd.Read4()
		o.Output = []byte(rd.ReadString(dataLen))
		fmt.Printf("Data: %s\n", string(o.Output))
		Entrys = append(Entrys, o)
	}

	if rd.err != nil {
		return nil, rd.err
	}

	return Entrys, nil
}
