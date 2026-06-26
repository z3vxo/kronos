package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/z3vxo/kronos/internal/database"
)



func NewFileManager(db *database.DB) *Manager {
	return &Manager{
		Uploads: make(map[string]*UploadTask),
		Downloads: make(map[Key]*DownloadTask),
		db:      db,
	}
}


func (m *Manager) InsertNewUploadFileTask(agentid string, taskid uint32, path, uuid string, hFile *os.File, rPath string) error {
	m.ulMu.Lock()
	defer m.ulMu.Unlock()

	m.Uploads[uuid] = &UploadTask{
		TaskID:     taskid,
		AgentID:    agentid,
		OnDiskFile: hFile,
		RemotePath: rPath,
		Status:     UploadStatusNotStarted,
	}

	return nil


}

func (m *Manager) InsertNewDownloadFileTask(agentid string, taskid uint32, filename string, fileID int64) error {
	m.dlMu.Lock()
	defer m.dlMu.Unlock()
	key := Key{
		AgentID: agentid,
		TaskID:  taskid,
	}

	if _, ok := m.Downloads[key]; ok {
		return ErrDownloadTaskExists
	}

	m.Downloads[key] = &DownloadTask{
		AgentID:      agentid,
		TaskID:       taskid,
		FileID:       fileID,
		OriginalPath: filename,
		Status:       StatusWaiting,
	}

	return nil
}

const UploadChunkSize = 512 * 1024

const (
	UploadChunked uint32 = iota + 1
	UploadNoChunked
	UploadDone
)

func (m *Manager) ReadUploadChunk(uuid string) ([]byte, bool, error) {
	m.ulMu.Lock()
	defer m.ulMu.Unlock()

	task, ok := m.Uploads[uuid]
	if !ok {
		return nil, false, fmt.Errorf("upload task not found: %s", uuid)
	}

	buf := make([]byte, UploadChunkSize)
	n, err := task.OnDiskFile.Read(buf)
	if n == 0 && err != nil {
		return nil, false, err
	}

	final := n < UploadChunkSize
	return buf[:n], final, nil
}

func (m *Manager) ProcessFileChunk(id string, taskID uint32, chunk Chunk) (ProcessResult, error) {
	m.dlMu.Lock()
	defer m.dlMu.Unlock()
	result := ProcessResult{}

	key := Key{AgentID: id, TaskID: taskID}

	task, ok := m.Downloads[key]
	if !ok {
		return ProcessResult{}, ErrDownloadTaskNotFound
	}

	if task.File == nil {
		result.Started = true
		home, err := os.UserHomeDir()
		if err != nil {
			return ProcessResult{}, err
		}

		agentDir := filepath.Join(home, ".kronos", "files", id)
		tmpDir := filepath.Join(agentDir, "tmp")
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			return ProcessResult{}, err
		}

		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%d.part", taskID))

		f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return ProcessResult{}, err
		}
		task.File = f
		task.TempPath = tmpPath
		task.Status = StatusOngoing
	}

	switch chunk.Status {
	case UploadChunked:
		if _, err := task.File.Write(chunk.Data); err != nil {
			task.Status = StatusFailed
			return ProcessResult{}, err
		}
		task.BytesSeen += uint64(len(chunk.Data))
		task.Status = StatusOngoing
		return result, nil

	case UploadNoChunked, UploadDone:
		if _, err := task.File.Write(chunk.Data); err != nil {
			task.Status = StatusFailed
			return ProcessResult{}, err
		}
		task.BytesSeen += uint64(len(chunk.Data))
		task.Status = StatusDone

		if err := task.File.Close(); err != nil {
			task.File = nil
			delete(m.Downloads, key)
			return ProcessResult{}, err
		}
		task.File = nil

		completedDir := filepath.Join(filepath.Dir(filepath.Dir(task.TempPath)), "completed")
		if err := os.MkdirAll(completedDir, 0755); err != nil {
			delete(m.Downloads, key)
			return ProcessResult{}, err
		}

		finalPath := filepath.Join(completedDir, filepath.Base(task.OriginalPath))
		if err := os.Rename(task.TempPath, finalPath); err != nil {
			delete(m.Downloads, key)
			return ProcessResult{}, err
		}

		task.FinalPath = finalPath
		delete(m.Downloads, key)
		m.db.UpdateFileDone(task.FileID, 1, finalPath, task.BytesSeen)
		
		result.Done = true
		result.FinalPath = finalPath
		result.BytesSeen = task.BytesSeen
		return result, nil
	default:
		return ProcessResult{}, ErrUnknownDownloadStatus
	}
}
