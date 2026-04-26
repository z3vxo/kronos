package database

import (
	"fmt"
	"time"
)

func (db *DB) InsertCommand(cmdType int, taskid, guid, param1, param2 string) error {

	query := `INSERT INTO commands(guid, command_type, task_id, param_1, param_2,executed, tasked_at) VALUES(?, ?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, guid, cmdType, taskid, param1, param2, 0, time.Now().Unix())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func (db *DB) GetTasks(id string) ([]Task, error) {
	q := `SELECT guid, command_type, task_id, param_1, param_2 FROM commands WHERE guid = ? AND executed = 0 LIMIT 3`
	rows, err := db.conn.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	i := 1
	for rows.Next() {
		var t Task
		t.ID = i
		err := rows.Scan(&t.Guid, &t.CmdCode, &t.TaskID, &t.Param1, &t.Param2)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
		i += 1
	}

	return tasks, nil
}

func (db *DB) DeleteTask(guid, id string) error {
	query := `DELETE FROM commands WHERE task_id = ? AND guid = ?`

	_, err := db.conn.Exec(query, id, guid)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) ListTasks(guid string) ([]Task, error) {
	query := `SELECT task_id, command_type, param_1, param_2, tasked_at FROM commands WHERE guid = ? AND executed = 0`

	rows, err := db.conn.Query(query, guid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	i := 1
	for rows.Next() {
		var t Task
		t.ID = i
		err := rows.Scan(&t.TaskID, &t.CmdCode, &t.Param1, &t.Param2, &t.TaskedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
		i++
	}

	return tasks, nil

}
