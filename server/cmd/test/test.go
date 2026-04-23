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


type Task struct {
	CmdCode int32
	TaskID  int32
	Param1  string
	Param2  string
}

// number of params each cmd_code expects on the wire
var cmdParamCount = map[int32]int{
	0: 2, // ps
	1: 1, // cmd
	2: 2, // cat
	3: 1, // ls
	4: 1, // rm
	5: 2, // mv
}

func readString(r *bytes.Reader) (string, error) {
	var length uint32
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

func parseTasks(body []byte) ([]Task, error) {
	r := bytes.NewReader(body)

	var count int32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, fmt.Errorf("reading task count: %w", err)
	}

	tasks := make([]Task, 0, count)
	for i := 0; i < int(count); i++ {
		var t Task
		if err := binary.Read(r, binary.LittleEndian, &t.CmdCode); err != nil {
			return nil, fmt.Errorf("task %d: reading cmd code: %w", i, err)
		}
		if err := binary.Read(r, binary.LittleEndian, &t.TaskID); err != nil {
			return nil, fmt.Errorf("task %d: reading task id: %w", i, err)
		}
		n := cmdParamCount[t.CmdCode]
		if n >= 1 {
			if p1, err := readString(r); err == nil {
				t.Param1 = p1
			}
		}
		if n >= 2 {
			if p2, err := readString(r); err == nil {
				t.Param2 = p2
			}
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func buildOutputPayload(taskID int32, output string) []byte {
	var buf bytes.Buffer
	// outer cmd type: 2 = CMD_TYPE_OUTPUT
	binary.Write(&buf, binary.LittleEndian, int32(2))
	// output count
	binary.Write(&buf, binary.LittleEndian, int32(1))
	// task id
	binary.Write(&buf, binary.LittleEndian, taskID)
	// cmd type: 0 = forward to operator
	binary.Write(&buf, binary.LittleEndian, int32(0))
	// output len + data
	binary.Write(&buf, binary.LittleEndian, int32(len(output)))
	buf.WriteString(output)
	return buf.Bytes()
}

func main() {
	host := flag.String("host", "127.0.0.1", "listener host")
	port := flag.Int("port", 8080, "listener port")
	download := flag.String("download", "/ms/download", "GET endpoint")
	upload := flag.String("upload", "/ms/upload", "POST endpoint")
	proto := flag.String("proto", "https", "protocol: http or https")
	sendOutput := flag.Bool("output", false, "send test output instead of polling")
	flag.Parse()

	guid := "4551bcd5-ebf8-4110-8d39-79ddaf37fb54"
	baseURL := fmt.Sprintf("%s://%s:%d", *proto, *host, *port)

	client := &http.Client{}
	if *proto == "https" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	fmt.Printf("[*] Using guid=%s\n", guid)

	if *sendOutput {
		payload := buildOutputPayload(1, "uid=0(root) gid=0(root) groups=0(root)")
		req, err := http.NewRequest("POST", baseURL+*upload, bytes.NewReader(payload))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Build request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("X-Agent-ID", guid)
		req.Header.Set("Content-Type", "application/octet-stream")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Send output failed: %v\n", err)
			os.Exit(1)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("[+] Output sent: %s | %s\n", resp.Status, string(body))
		return
	}

	// Poll loop
	fmt.Printf("[*] Polling %s every 5s\n\n", baseURL+*download)
	for {
		time.Sleep(5 * time.Second)

		req, err := http.NewRequest("GET", baseURL+*download, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-] Build request: %v\n", err)
			continue
		}
		req.Header.Set("X-Agent-ID", guid)

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
}
