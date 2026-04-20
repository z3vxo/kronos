package database

import (
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (db *DB) ListAgents() ([]Agent, error) {

	qeuery := `SELECT code_name, username, hostname, external_ip, internal_ip, is_elevated, pid, process_path, windows_version, last_checkin FROM agents`

	rows, err := db.conn.Query(qeuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var AgentList []Agent

	for rows.Next() {
		var a Agent
		err := rows.Scan(&a.CodeName, &a.Username, &a.Hostname, &a.Ex_ip, &a.In_ip, &a.IsElevated, &a.Pid, &a.ProcPath, &a.WinVer, &a.LastSeen)
		if err != nil {
			return nil, err
		}
		AgentList = append(AgentList, a)
	}

	return AgentList, nil

}

func (db *DB) ResolveCodename(name string) (string, error) {
	query := `SELECT guid FROM agents WHERE code_name = ?`
	var guid string
	err := db.conn.QueryRow(query, name).Scan(&guid)

	if err != nil {
		return "", err
	}

	return guid, nil
}

func (db *DB) InsertAgent(guid, codeName, User, Host, InIP, ExIP, ProcPath, WinVer string, Pid int32, IsElev byte) error {
	query := `INSERT INTO agents(guid, code_name,
	  						username, hostname,
							external_ip, internal_ip,
							is_elevated, pid, process_path,
							windows_version, session_key, last_checkin) VALUES(
							?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, guid, codeName, User, Host, ExIP, InIP, IsElev, Pid, ProcPath, WinVer, "32324234", time.Now().UTC().Unix())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
