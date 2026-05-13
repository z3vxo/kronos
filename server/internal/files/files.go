package files

import (
	"errors"
	"fmt"
	"github.com/z3vxo/kronos/internal/database"
	"os"
	"path/filepath"
	"sync"
)

type Status uint32

const (
	StatusWaiting Status = iota + 1
	StatusOngoing
	StatusDone
	StatusFailed
)

type Manager struct {
	mu      sync.Mutex
	uploads map[Key]*UploadTask
	db      *database.DB
}

type Key struct {
	AgentID string
	TaskID  uint32
}

type UploadTask struct {
	AgentID      string
	TaskID       uint32
	TempPath     string
	FinalPath    string
	OriginalPath string
	File         *os.File
	BytesSeen    uint64
	TotalSize    uint64
	Status       Status
}

type Chunk struct {
	Status uint32
	Data   []byte
}

type ProcessResult struct {
	Started   bool
	Done      bool
	FinalPath string
	BytesSeen uint64
}

var ErrUploadTaskExists = errors.New("upload task already exists")
var ErrUploadTaskNotFound = errors.New("upload task not found")
var ErrUnknownUploadStatus = errors.New("unknown upload status")

func NewFileManager(db *database.DB) *Manager {
	return &Manager{
		uploads: make(map[Key]*UploadTask),
		db:      db,
	}
}

func (m *Manager) InsertNewFileTask(agentid string, taskid uint32, filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := Key{
		AgentID: agentid,
		TaskID:  taskid,
	}

	if _, ok := m.uploads[key]; ok {
		return ErrUploadTaskExists
	}

	m.uploads[key] = &UploadTask{
		AgentID:      agentid,
		TaskID:       taskid,
		OriginalPath: filename,
		Status:       StatusWaiting,
	}

	return nil
}

const (
	UploadChunked uint32 = iota + 1
	UploadNoChunked
	UploadDone
)

func (m *Manager) ProcessFileChunk(id string, taskID uint32, chunk Chunk) (ProcessResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := ProcessResult{}

	key := Key{AgentID: id, TaskID: taskID}

	task, ok := m.uploads[key]
	if !ok {
		return ProcessResult{}, ErrUploadTaskNotFound
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
			delete(m.uploads, key)
			return ProcessResult{}, err
		}
		task.File = nil

		completedDir := filepath.Join(filepath.Dir(filepath.Dir(task.TempPath)), "completed")
		if err := os.MkdirAll(completedDir, 0755); err != nil {
			delete(m.uploads, key)
			return ProcessResult{}, err
		}

		finalPath := filepath.Join(completedDir, filepath.Base(task.OriginalPath))
		if err := os.Rename(task.TempPath, finalPath); err != nil {
			delete(m.uploads, key)
			return ProcessResult{}, err
		}

		task.FinalPath = finalPath
		delete(m.uploads, key)
		if err := m.db.InsertFile(id, finalPath, task.BytesSeen); err != nil {
			return ProcessResult{}, err
		}
		result.Done = true
		result.FinalPath = finalPath
		result.BytesSeen = task.BytesSeen
		return result, nil
	default:
		return ProcessResult{}, ErrUnknownUploadStatus
	}
}
