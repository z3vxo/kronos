package database

import (
	"fmt"
)

func (db *DB) InsertFile(agentID string, onDiskPath string, size uint64) error {
	q := "INSERT INTO files(agentid, onDiskPath, size) VALUES(?, ?, ?);"
	if _, err := db.conn.Exec(q, agentID, onDiskPath, size); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
