package teamserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/z3vxo/kronos/internal/httputil"
)

func (ts *TeamServer) FilesSyncHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "code")
	diskPath, err := ts.db.GetFilePath(id)
	if err != nil {
		fmt.Println("failed finding file 1")

		httputil.SendJSONError(w, "Could not find File", http.StatusNotFound)
		return
	}
	f, err := os.Open(diskPath)
	if err != nil {
		fmt.Println("failed finding file 2")
		httputil.SendJSONError(w, "Could not open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(diskPath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}

func (ts *TeamServer) FilesListHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ts.db.GetFiles()
	if err != nil {
		httputil.SendJSONError(w, "Failed Retreiving files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&files)
}


func (ts *TeamServer) DownloadTaskHandler(w http.ResponseWriter, r *http.Request) {
	var cmd TaskEntry
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		httputil.SendJSONError(w, "Error decoding json", http.StatusInternalServerError)
		return
	}

	taskID := GenTaskID()

	fileID, err := ts.db.InsertFile(cmd.Guid, "", 0, "download")
	if err != nil {
		httputil.SendJSONError(w, "failed inserting file record", http.StatusInternalServerError)
		return
	}

	if err := ts.FileMgr.InsertNewFileTask(cmd.Guid, taskID, cmd.Param1, fileID); err != nil {
		httputil.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ts.db.InsertCommand(cmd.Cmd_type, taskID, cmd.Guid, cmd.Param1, cmd.Param2); err != nil {
		httputil.SendJSONError(w, "failed inserting command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}
