package database

import (
	"fmt"
	"time"
)

func (db *DB) InsertCommand(cmdType int, taskid uint32, guid, param1, param2 string, params []byte) error {

	q := `INSERT INTO commands(guid, task_id, command_type, param_1, param_2, params, executed, tasked_at) VALUES(?,?,?,?,?,?,?,?) `

	_, err := db.conn.Exec(q, guid, taskid, cmdType, param1, param2, params, 0, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil

}


func (db *DB) GetTasks(id string) ([]Task, error) {
	q := "SELECT guid, command_type, task_id, params, param_1, param_2 FROM commands WHERE guid = ? AND executed = 0 LIMIT 3"
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
		err := rows.Scan(&t.Guid, &t.CmdCode, &t.TaskID, &t.Params, &t.Param1, &t.Param2)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
		i += 1
	}

	return tasks, nil
}

func (db *DB) DeleteTask(guid string, id string) error {
	fmt.Printf("Deleting %s\n", guid)
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


func (db *DB) GetSingleTask(guid, taskID string) (Task, error) {
	var t Task
	query := `SELECT task_id, command_type, param_1, param_2, tasked_at FROM commands WHERE guid = ? AND task_id = ?`
	err := db.conn.QueryRow(query, guid, taskID).Scan(&t.TaskID, &t.CmdCode, &t.Param1, &t.Param2, &t.TaskedAt)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}