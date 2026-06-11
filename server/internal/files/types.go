package files

import (
	"errors"
	"os"
	"sync"

	"github.com/z3vxo/kronos/internal/database"
)

type Status uint32

const (
	StatusWaiting Status = iota + 1
	StatusOngoing
	StatusDone
	StatusFailed
)

type Manager struct {
	dlMu    sync.Mutex
	ulMu    sync.Mutex
	Downloads   map[Key]*DownloadTask
	Uploads     map[string]*UploadTask
	db      *database.DB
}

type UploadTask struct {
	TaskID uint32
	AgentID string
	OnDiskFile *os.File
	RemotePath string
	Status Status
}

type Key struct {
	AgentID string
	TaskID  uint32
}

type DownloadTask struct {
	AgentID      string
	TaskID       uint32
	FileID       int64
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

var ErrDownloadTaskExists = errors.New("download task already exists")
var ErrDownloadTaskNotFound = errors.New("download task not found")
var ErrUnknownDownloadStatus = errors.New("unknown download status")


const (
	UploadStatusNotStarted Status = 1
	UploadStatusOngoing    Status = 2
	UploadStatusDone       Status = 3
)