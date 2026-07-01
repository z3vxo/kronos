package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func GetDbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.kronos/database/kronos_db.sql", home), nil
}

func NewDB() (*DB, error) {
	dbPath, err := GetDbPath()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	d := &DB{conn: db}
	err = InitDB(d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func InitDB(db *DB) error {

	pragmans := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
	}

	for _, p := range pragmans {
		if _, err := db.conn.Exec(p); err != nil {
			fmt.Println("pragmas")
			return err
		}
	}

	return SetupDB(db)
}

func SetupDB(db *DB) error {

	agents := `CREATE TABLE IF NOT EXISTS agents (
		guid 			INTEGER NOT NULL,
		code_name 		TEXT NOT NULL DEFAULT '',
		username 		TEXT NOT NULL DEFAULT '',
		hostname 		TEXT NOT NULL DEFAULT '',
		external_ip	 	TEXT NOT NULL DEFAULT '',
		internal_ip 	TEXT NOT NULL DEFAULT '',
		is_elevated 	BOOLEAN NOT NULL DEFAULT 0,
		arch			INTEGER NOT NULL DEFAULT 0,
		pid 			INTEGER NOT NULL DEFAULT 0,
		tid      		INTEGER NOT NULL DEFAULT 0,
		ppid			INTEGER NOT NULL DEFAULT 0,
		process_path 	TEXT NOT NULL DEFAULT '',
		windows_version TEXT NOT NULL DEFAULT '',
		session_key    	INTEGER NOT NULL,
		sleep           INTEGER NOT NULL DEFAULT 0,
		jitter          INTEGER NOT NULL DEFAULT 0,
		registered      BOOLEAN NOT NULL DEFAULT 0,
		last_checkin    INTEGER NOT NULL DEFAULT 0,
		registration_date INTEGER NOT NULL DEFAULT 0);`

	_, err := db.conn.Exec(agents)
	if err != nil {
		fmt.Println(err)
		fmt.Println("agents")
		return err
	}

	commands := `CREATE TABLE IF NOT EXISTS commands (
		guid INTEGER NOT NULL,
		command_type INTEGER NOT NULL,
		task_id      INTEGER NOT NULL,
		param_1      TEXT NOT NULL,
		param_2      TEXT NOT NULL,
		executed     BOOLEAN NOT NULL,
		tasked_at    INTEGER NOT NULL);`

	_, err = db.conn.Exec(commands)
	if err != nil {

		return err
	}

	listeners := `CREATE TABLE IF NOT EXISTS listeners (
		id INTEGER PRIMARY KEY,
		guid TEXT NOT NULL,
		port INTEGER NOT NULL,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		certType BOOLEAN NOT NULL,
		protocol TEXT NOT NULL,
		status INTEGER NOT NULL);
		`
	_, err = db.conn.Exec(listeners)
	if err != nil {
		return err
	}

	files := `CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY,
		status INTEGER,
		agentid TEXT NOT NULL,
		onDiskPath TEXT NOT NULL,
		size INTEGER NOT NULL,
	    type TEXT NOT NULL);`
	_, err = db.conn.Exec(files)
	if err != nil {
		return err
	}

	profiles := `CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		GetEndpoint TEXT NOT NULL,
		PostEndpoint TEXT NOT NULL,
		jitter INTEGER NOT NULL,
		sleep INTEGER NOT NULL,
		SleepObf BOOLEAN,
		HeapObf BOOLEAN,
		StackSpoof BOOLEAN,
		Syscall INTEGER
	);`

	_, err = db.conn.Exec(profiles)
	if err != nil {
		return err
	}

	profile_domains := `CREATE TABLE IF NOT EXISTS profile_domains (
		id INTEGER NOT NULL REFERENCES profiles(id),
		domain TEXT NOT NULL,
		port INTEGER NOT NULL,
		isHttps BOOLEAN
	);`
	_, err = db.conn.Exec(profile_domains)
	if err != nil {
		return err
	}
	profile_headers := `CREATE TABLE IF NOT EXISTS profile_headers (
		id INTEGER NOT NULL REFERENCES profiles(id),
		Key TEXT NOT NULL,
		Value TEXT NOT NULL
	);`
	_, err = db.conn.Exec(profile_headers)
	if err != nil {
		return err
	}

	return nil
}
