package teamserver

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/z3vxo/kronos/internal/database"
	"github.com/z3vxo/kronos/internal/httputil"
	"github.com/z3vxo/kronos/internal/bytemgr"
)

const TASK_DOWNLOAD = 12

func GenTaskID() uint32 {
	var b [4]byte
	rand.Read(b[:])
	return binary.LittleEndian.Uint32(b[:])
}

func (ts *TeamServer) CommandNewHandler(w http.ResponseWriter, r *http.Request) {
	var cmd TaskEntry
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.SendJSONError(w, "Error decoding json", http.StatusInternalServerError)
		return
	}

	taskID := GenTaskID()

	Buf, err := bytemgr.CraftCmdFormat(uint32(cmd.Cmd_type), taskID, cmd.Param1, cmd.Param2, cmd.DataType)
	if err != nil {
		httputil.SendJSONError(w, "failed crafting format", http.StatusInternalServerError)
		return
	}


	err = ts.db.InsertCommand(cmd.Cmd_type, taskID, cmd.Guid, cmd.Param1, cmd.Param2, Buf)
	if err != nil {
		httputil.SendJSONError(w, "failed inserting command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (ts *TeamServer) CommandDeleteHandler(w http.ResponseWriter, r *http.Request) {
	guid := chi.URLParam(r, "guid")
	taskID := chi.URLParam(r, "taskID")
	if guid == "" || taskID == "" {
		httputil.SendJSONError(w, "missing guid AND/OR taskID", http.StatusBadRequest)
		return
	}

	if err := ts.db.DeleteTask(guid, taskID); err != nil {
		httputil.SendJSONError(w, "Failed Deleting task", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})

}

func (ts *TeamServer) ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	guid := chi.URLParam(r, "guid")
	if guid == "" {
		httputil.SendJSONError(w, "missing guid", http.StatusBadRequest)
		return
	}

	tasks, err := ts.db.ListTasks(guid)
	if err != nil {
		httputil.SendJSONError(w, "database error, failed loading tasks", http.StatusInternalServerError)
		return
	}

	for i := range tasks {
		tasks[i].CmdName = bytemgr.CmdNames[tasks[i].CmdCode]
	}

	payload := database.TaskEntrys{Total: len(tasks), Tasks: tasks}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&payload)

}


func (ts *TeamServer) GetHistoryOutputHandler(w http.ResponseWriter, r *http.Request) {
	guid := chi.URLParam(r, "guid")
	taskID := chi.URLParam(r, "taskID")
	if guid == "" || taskID == "" {
		httputil.SendJSONError(w, "missing guid or taskID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		httputil.SendJSONError(w, "invalid taskID", http.StatusBadRequest)
		return
	}

	output, err := ts.Logger.GetOutput(guid, uint32(id))
	if err != nil {
		httputil.SendJSONError(w, "output not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func (ts *TeamServer) ListHistoryHandler(w http.ResponseWriter, r *http.Request) {
	guid := chi.URLParam(r, "guid")

	data, err := ts.Logger.GetHistory(guid)
	if err != nil {
		httputil.SendJSONError(w, "Failed getting history", http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&data)


}
