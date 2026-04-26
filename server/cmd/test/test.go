package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// number of params each cmd_code writes on the wire (server omits empty params)
var cmdParamCount = map[int32]int{
	0: 2, // ps
	1: 1, // cmd
	2: 2, // cat
	3: 1, // ls
	4: 1, // rm
	5: 2, // mv
}

func writeI32(w *bytes.Buffer, v int32) { binary.Write(w, binary.LittleEndian, v) }
func writeI16(w *bytes.Buffer, v int16) { binary.Write(w, binary.LittleEndian, v) }
func writeStr(w *bytes.Buffer, s string) {
	writeI32(w, int32(len(s)))
	w.WriteString(s)
}

func buildRegisterPayload(guid, user, host, ip, path string, pid, ppid int32, elev, arch byte, minor, major, build int16) []byte {
	var buf bytes.Buffer
	writeI32(&buf, 1) // CMD_TYPE_REGISTER
	writeStr(&buf, guid)
	writeStr(&buf, user)
	writeStr(&buf, host)
	writeStr(&buf, ip)
	writeStr(&buf, path)
	writeI32(&buf, pid)
	writeI32(&buf, ppid)
	buf.WriteByte(elev)
	buf.WriteByte(arch)
	writeI16(&buf, minor)
	writeI16(&buf, major)
	writeI16(&buf, build)
	return buf.Bytes()
}

func buildOutputPayload(taskID int32, output string) []byte {
	var buf bytes.Buffer
	writeI32(&buf, 2) // CMD_TYPE_OUTPUT
	writeI32(&buf, 1) // count
	writeI32(&buf, taskID)
	writeI32(&buf, 0) // type: 0 = forward to operator
	writeI32(&buf, int32(len(output)))
	buf.WriteString(output)
	return buf.Bytes()
}

func readString(r *bytes.Reader) (string, error) {
	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

type Task struct {
	CmdCode int32
	TaskID  int32
	Param1  string
	Param2  string
}

func parseTasks(body []byte) ([]Task, error) {
	r := bytes.NewReader(body)
	var count int32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading task count: %w", err)
	}
	tasks := make([]Task, 0, count)
	for i := range int(count) {
		var t Task
		if err := binary.Read(r, binary.LittleEndian, &t.CmdCode); err != nil {
			return nil, fmt.Errorf("task %d: reading cmd code: %w", i, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &t.TaskID); err != nil {
			return nil, fmt.Errorf("task %d: reading task id: %w", i, err)
		}
		n := cmdParamCount[t.CmdCode]
		if n >= 1 {
			t.Param1, _ = readString(r)
		}
		if n >= 2 {
			t.Param2, _ = readString(r)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func post(client *http.Client, url, guid string, body []byte) (int, string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	if guid != "" {
		req.Header.Set("X-Agent-ID", guid)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b), nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [global flags] <register|output|task> [flags]\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "Global flags:")
	fmt.Fprintln(os.Stderr, "  -host string    listener host (default 127.0.0.1)")
	fmt.Fprintln(os.Stderr, "  -port int       listener port (default 8080)")
	fmt.Fprintln(os.Stderr, "  -proto string   http or https (default https)")
	fmt.Fprintln(os.Stderr, "  -guid string    agent GUID")
	fmt.Fprintln(os.Stderr, "  -download path  GET endpoint (default /ms/download)")
	fmt.Fprintln(os.Stderr, "  -upload path    POST endpoint (default /ms/upload)")
	fmt.Fprintln(os.Stderr, "\nSubcommands:")
	fmt.Fprintln(os.Stderr, "  register  send registration payload to upload endpoint")
	fmt.Fprintln(os.Stderr, "  output    send task output to upload endpoint")
	fmt.Fprintln(os.Stderr, "  task      poll download endpoint for pending tasks")
	os.Exit(1)
}

func main() {
	host := flag.String("host", "127.0.0.1", "listener host")
	port := flag.Int("port", 8080, "listener port")
	download := flag.String("download", "/ms/download", "GET endpoint")
	upload := flag.String("upload", "/ms/upload", "POST endpoint")
	proto := flag.String("proto", "https", "protocol: http or https")
	guid := flag.String("guid", "4551bcd5-ebf8-4110-8d39-79ddaf37fb54", "agent GUID")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	client := &http.Client{}
	if *proto == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	baseURL := fmt.Sprintf("%s://%s:%d", *proto, *host, *port)
	sub := flag.Args()[0]
	subArgs := flag.Args()[1:]

	switch sub {
	case "register":
		fs := flag.NewFlagSet("register", flag.ExitOnError)
		user := fs.String("user", "testuser", "username")
		hostname := fs.String("hostname", "TESTHOST", "hostname")
		ip := fs.String("ip", "192.168.1.100", "internal IP")
		path := fs.String("path", `C:\Windows\System32\cmd.exe`, "process path")
		pid := fs.Int("pid", 1234, "PID")
		ppid := fs.Int("ppid", 1000, "PPID")
		elev := fs.Int("elev", 0, "is elevated (0 or 1)")
		arch := fs.Int("arch", 1, "arch: 1=x64, 0=x86")
		minor := fs.Int("minor", 0, "OS minor version")
		major := fs.Int("major", 10, "OS major version")
		build := fs.Int("build", 19045, "OS build version")
		fs.Parse(subArgs)

		payload := buildRegisterPayload(
			*guid, *user, *hostname, *ip, *path,
			int32(*pid), int32(*ppid),
			byte(*elev), byte(*arch),
			int16(*minor), int16(*major), int16(*build),
		)
		status, body, err := post(client, baseURL+*upload, "", payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Register failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[+] Register: %d | %s\n", status, body)

	case "output":
		fs := flag.NewFlagSet("output", flag.ExitOnError)
		taskID := fs.Int("task-id", 1, "task ID")
		data := fs.String("data", "uid=0(root) gid=0(root) groups=0(root)", "output string")
		fs.Parse(subArgs)

		payload := buildOutputPayload(int32(*taskID), *data)
		status, body, err := post(client, baseURL+*upload, *guid, payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Output failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[+] Output sent: %d | %s\n", status, body)

	case "task":
		fmt.Printf("[*] Polling %s every 5s (guid=%s)\n\n", baseURL+*download, *guid)
		for {
			time.Sleep(5 * time.Second)
			req, err := http.NewRequest("GET", baseURL+*download, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[-] Build request: %v\n", err)
				continue
			}
			req.Header.Set("X-Agent-ID", *guid)
			resp, err := client.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[-] Checkin failed: %v\n", err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			switch resp.StatusCode {
			case http.StatusNoContent:
				fmt.Printf("[*] No tasks\n")
			case http.StatusOK:
				tasks, err := parseTasks(body)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[-] Parse failed: %v\n", err)
					continue
				}
				fmt.Printf("[+] %d task(s):\n", len(tasks))
				for i, t := range tasks {
					fmt.Printf("    [%d] cmd_code=%d task_id=%d param1=%q param2=%q\n", i, t.CmdCode, t.TaskID, t.Param1, t.Param2)
				}
			default:
				fmt.Printf("[!] Unexpected status: %s | %s\n", resp.Status, string(body))
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "[-] Unknown subcommand: %s\n\n", sub)
		usage()
	}
}
