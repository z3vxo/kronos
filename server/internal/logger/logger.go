package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type HistoryEntry struct {
	CmdCode    int    `json:"cmd_code"`
	CmdName    string `json:"cmd_name"`
	TaskID     uint32 `json:"task_id"`
	Param1     string `json:"param_1"`
	Param2     string `json:"param_2"`
	TaskedAt   int64  `json:"tasked_at"`
	FinishedAt int64  `json:"finished_at"`
}

type History struct {
	Total int            `json:"total"`
	Entry []HistoryEntry `json:"history"`
}

type Logger struct {
	operatorOps *slog.Logger
	tasksDir    string
}

func New() (*Logger, error) {
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, ".kronos", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filepath.Join(logsDir, "operators.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		operatorOps: slog.New(slog.NewJSONHandler(f, nil)),
		tasksDir:    filepath.Join(logsDir, "tasks"),
	}, nil
}

func (l *Logger) LogOperatorOp(msg, event string, args ...any) {
	l.operatorOps.Info(msg, append([]any{"event", event}, args...)...)
}

func (l *Logger) LogMetaAndOutput(guid string, taskID uint32, output []byte, args ...any) error {
	dir := filepath.Join(l.tasksDir, guid, "output")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	metaPath := filepath.Join(l.tasksDir, guid, "meta.jsonl")
	f, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	slog.New(slog.NewJSONHandler(f, nil)).Info("task_complete", append([]any{
		"task_id", taskID,
		"finished_at", time.Now().Unix(),
	}, args...)...)

	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.txt", taskID)), output, 0644)
}

func (l *Logger) GetHistory(guid string) (History, error) {
	path := filepath.Join(l.tasksDir, guid, "meta.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return History{}, nil
		}
		return History{}, err
	}
	defer f.Close()

	var history History
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var h HistoryEntry
		if err := json.Unmarshal(line, &h); err != nil {
			return History{}, err
		}
		history.Entry = append(history.Entry, h)
	}
	if err := scanner.Err(); err != nil {
		return History{}, err
	}
	history.Total = len(history.Entry)
	return history, nil
}

func (l *Logger) GetOutput(guid string, taskID uint32) ([]byte, error) {
	path := filepath.Join(l.tasksDir, guid, "output", fmt.Sprintf("%d.txt", taskID))
	return os.ReadFile(path)
}
