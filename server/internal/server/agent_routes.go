package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
)

const (
	CMD_TYPE_REGISTER = 1
	CMD_TYPE_OUTPUT   = 2
)

func (h *AgentHandler) AgentCheckInHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS"))
}

func (h *AgentHandler) AgentUploadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(body)
	var cmdType int32
	if err := binary.Read(reader, binary.LittleEndian, &cmdType); err != nil {
		http.Error(w, "failed reading cmd type", http.StatusInternalServerError)
		return
	}

	switch cmdType {
	case CMD_TYPE_REGISTER:
		fmt.Println("[+] Register hit!")
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		err := h.HandleClientRegister(ip, reader)
		if err != nil {
			http.Error(w, "failed getting data", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SUCCESS"))

	}
}
