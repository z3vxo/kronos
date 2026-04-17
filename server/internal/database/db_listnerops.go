package database

func (db *DB) InsertListener(port int, id string) error {
	query := `INSERT INTO listeners(port, guid, status) VALUES(?, ?, ?)`

	_, err := db.conn.Exec(query, port, id, "running")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) DeleteListener(id string) error {
	query := `DELETE FROM listeners WHERE guid = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
