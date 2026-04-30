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

func (r *Reader) Read4() int32 {
	if r.err != nil {
		return 0
	}
	var val int32
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

func (r *Reader) ReadString(len int32) string {
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
	Guid       int32
	User       string
	Host       string
	InternaIP  string
	ExternalIP string
	ProcPath   string
	Pid        int32
	Tid        int32
	Ppid       int32
	IsElev     byte
	Arch       byte
	Minor      int32
	Major      int32
	Build      int32
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
	fmt.Println(guid)
	fmt.Println(Username)
	fmt.Println(Hostname)
	fmt.Println(InternalIP)
	fmt.Println(ProcessPath)
	fmt.Println(Pid)
	fmt.Println(Tid)
	fmt.Println(PPid)
	fmt.Println(IsElev)
	fmt.Println(Arch)
	fmt.Println(Minor)
	fmt.Println(Major)
	fmt.Println(BuildVer)

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
	Type   int32
	TaskID int32
	Output []byte
}

// Looped
/* [OUTPUT COUNT] 4 BYTES
 * ->
 * [Task ID] 4 BYTES
 * ->
 * [CMD TYPE] 4 BYTES | 0 == Server does nothing, 1 == File Content(uses task ID for Map lookup)
 * ->
 * [OUTPUT LEN]  4 BYTES
 * [OUTPUT DATA] N BYTES
 */

func ParseClientOutput(r *bytes.Reader) ([]OutputEntrys, error) {

	rd := Reader{r: r}
	var Entrys []OutputEntrys
	Count := rd.Read4()

	for range Count {
		var o OutputEntrys
		o.TaskID = rd.Read4()
		o.Type = rd.Read4()
		OutputLen := rd.Read4()
		o.Output = []byte(rd.ReadString(OutputLen))
		if rd.err != nil {
			return nil, rd.err
		}
		Entrys = append(Entrys, o)
	}

	return Entrys, nil
}
