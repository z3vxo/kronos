package bytemgr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
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

func (r *Reader) Read8() uint64 {
	if r.err != nil {
		return 0
	}
	var val uint64
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

func (r *Reader) Remaining() int {
	return r.r.Len()
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
		fmt.Println(rd.err)
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

func ParseClientOutput(r *bytes.Reader) ([]OutputEntrys, error) {
	rd := Reader{r: r}
	var entrys []OutputEntrys
	for rd.Remaining() > 0 {
		if rd.Remaining() < 12 {
			return nil, io.ErrUnexpectedEOF
		}

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
		if TaskType == 6 {
			o.FileOutput = ParseFileOutput(&rd)
			if rd.err != nil {
				return nil, rd.err
			}
			entrys = append(entrys, o)
			continue
		}
		if parser, ok := taskOutputParsers[TaskType]; ok {
			o.Output = []byte(parser(&rd))
			if rd.err != nil {
				return nil, rd.err
			}
			entrys = append(entrys, o)
			continue
		}

		HasData := rd.Read4()
		if HasData > 1 {
			SuccessString := SuccessMap[HasData]
			o.Output = []byte(SuccessString)
			entrys = append(entrys, o)
			continue
		}
		OutputLen := rd.Read4()
		o.Output = []byte(rd.ReadString(OutputLen))
		if rd.err != nil {
			return nil, rd.err
		}
		entrys = append(entrys, o)
	}

	return entrys, nil
}

func ParseFileOutput(r *Reader) *FileOutput {
	status := FileStatus(r.Read4())
	if r.err != nil {
		return nil
	}

	dataLen := r.Read4()
	if r.err != nil {
		return nil
	}

	data := []byte(r.ReadString(dataLen))
	if r.err != nil {
		return nil
	}

	return &FileOutput{
		Status:  status,
		DataLen: dataLen,
		Data:    data,
	}
}

func ParsePSOutput(r *Reader) string {
	var b strings.Builder

	for {
		val := r.Read4()
		if val == LS_END {
			return b.String()
		}

		ProcessName := r.ReadString(val)
		UserLen := r.Read4()
		UserStr := r.ReadString(UserLen)
		PID := r.Read4()
		fmt.Fprintf(&b, "%-40s %-25s %d\n", ProcessName, UserStr, PID)
	}
	fmt.Println(b.String())
	return b.String()

}

func ParseLSOutput(r *Reader) string {
	var b strings.Builder

	for {
		val := r.Read4()
		if val == LS_END {
			return b.String()
		}
		EntryStr := r.ReadString(val)
		TypeLen := r.Read4()
		TypeStr := r.ReadString(TypeLen)
		FileSize := r.Read8()
		fmt.Fprintf(&b, "%-40s %-15s %d\n", EntryStr, TypeStr, FileSize)
	}
	fmt.Println(b.String())
	return b.String()
}

func ParsePRIVOutput(r *Reader) string {
	var b strings.Builder

	NameLen := r.Read4()
	Name := r.ReadString(NameLen)
	DomainLen := r.Read4()
	Domain := r.ReadString(DomainLen)
	SidLen := r.Read4()
	Sid := r.ReadString(SidLen)
	UserName := strings.ToLower(fmt.Sprintf("%s\\%s", Domain, Name))

	fmt.Fprintf(&b, "USER INFORMATION\n")
	fmt.Fprintf(&b, "----------------\n\n")
	fmt.Fprintf(&b, "%-22s %s\n", "User Name", "SID")
	fmt.Fprintf(&b, "%-22s %s\n", strings.Repeat("=", 22), strings.Repeat("=", 46))
	fmt.Fprintf(&b, "%-22s %s\n\n", UserName, Sid)

	fmt.Fprintf(&b, "PRIVILEGES INFORMATION\n")
	fmt.Fprintf(&b, "----------------------\n\n")
	fmt.Fprintf(&b, "%-29s %-36s %s\n", "Privilege Name", "Description", "State")
	fmt.Fprintf(&b, "%-29s %-36s %s\n", strings.Repeat("=", 29), strings.Repeat("=", 36), strings.Repeat("=", 8))

	for {
		PrivNameLen := r.Read4()
		if r.err != nil || PrivNameLen == LS_END {
			return b.String()
		}

		PrivName := r.ReadString(PrivNameLen)
		PrivStatus := r.Read4()
		Status, ok := privilegeStatusNames[PrivStatus]
		if !ok {
			Status = fmt.Sprintf("Unknown (%d)", PrivStatus)
		}
		Description := privilegeDescriptions[PrivName]

		fmt.Fprintf(&b, "%-29s %-36s %s\n", PrivName, Description, Status)
	}
}
