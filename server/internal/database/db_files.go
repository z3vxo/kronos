package database

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
)

type FileEntry struct {
	ID      uint32 `json:"id"`
	AgentID string `json:"agent_id"`
	Name    string `json:"filename"`
	Size    uint64 `json:"size"`
}

type Files struct {
	Total int         `json:"total"`
	Files []FileEntry `json:"files"`
}

func (db *DB) InsertFile(agentID string, onDiskPath string, size uint64) error {
	q := "INSERT INTO files(agentid, onDiskPath, size) VALUES(?, ?, ?);"
	if _, err := db.conn.Exec(q, agentID, onDiskPath, size); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (db *DB) GetFiles() (Files, error) {
	q := "SELECT id, agentid, onDiskPath, size FROM files;"
	rows, err := db.conn.Query(q)
	if err != nil {
		return Files{}, err
	}
	defer rows.Close()

	var entries []FileEntry
	for rows.Next() {
		var e FileEntry
		var path string
		if err := rows.Scan(&e.ID, &e.AgentID, &path, &e.Size); err != nil {
			return Files{}, err
		}
		e.Name = filepath.Base(path)
		entries = append(entries, e)
	}

	return Files{Total: len(entries), Files: entries}, nil
}

func (db *DB) GetFilePath(id string) (string, error) {
	I, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	}
	var path string
	if err = db.conn.QueryRow("SELECT onDiskPath FROM files WHERE id = ?", I).Scan(&path); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no file found for ID")
		}
		return "", err
	}
	return path, nil
}
