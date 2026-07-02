package database

import "fmt"

func (db *DB) InsertProfile(p Profile, id uint32, Key []byte) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO profiles (name, payload_id, Key, GetEndpoint, PostEndpoint, jitter, sleep, SleepObf, HeapObf, StackSpoof, Syscall)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, id, Key, p.Get, p.Post, p.Jitter, p.Sleep, p.SleepObf, p.HeapObf, p.StackSpoof, p.Syscall)
	if err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}

	profileID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	for _, d := range p.Domains {
		_, err := tx.Exec(`INSERT INTO profile_domains (id, domain, port, isHttps) VALUES (?, ?, ?, ?)`,
			profileID, d.Domain, d.Port, d.IsHttps)
		if err != nil {
			return fmt.Errorf("insert domain: %w", err)
		}
	}

	for _, h := range p.Headers {
		_, err := tx.Exec(`INSERT INTO profile_headers (id, Key, Value) VALUES (?, ?, ?)`,
			profileID, h.Key, h.Value)
		if err != nil {
			return fmt.Errorf("insert header: %w", err)
		}
	}

	return tx.Commit()
}

func (db *DB) ListProfiles() ([]Profile, error) {
	rows, err := db.conn.Query(`SELECT id, name, GetEndpoint, PostEndpoint, jitter, sleep, SleepObf, HeapObf, StackSpoof, Syscall FROM profiles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		err := rows.Scan(&p.ID, &p.Name, &p.Get, &p.Post, &p.Jitter, &p.Sleep, &p.SleepObf, &p.HeapObf, &p.StackSpoof, &p.Syscall)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}

	for i := range profiles {
		domRows, err := db.conn.Query(`SELECT domain, port, isHttps FROM profile_domains WHERE id = ?`, profiles[i].ID)
		if err != nil {
			return nil, err
		}
		for domRows.Next() {
			var d ProfileDomain
			if err := domRows.Scan(&d.Domain, &d.Port, &d.IsHttps); err != nil {
				domRows.Close()
				return nil, err
			}
			profiles[i].Domains = append(profiles[i].Domains, d)
		}
		domRows.Close()

		hdrRows, err := db.conn.Query(`SELECT Key, Value FROM profile_headers WHERE id = ?`, profiles[i].ID)
		if err != nil {
			return nil, err
		}
		for hdrRows.Next() {
			var h ProfileHeader
			if err := hdrRows.Scan(&h.Key, &h.Value); err != nil {
				hdrRows.Close()
				return nil, err
			}
			profiles[i].Headers = append(profiles[i].Headers, h)
		}
		hdrRows.Close()
	}

	return profiles, nil
}

func (db *DB) GetProfile(name string) (Profile, error) {
	var p Profile
	err := db.conn.QueryRow(`SELECT id, payload_id, Key, name, GetEndpoint, PostEndpoint, jitter, sleep, SleepObf, HeapObf, StackSpoof, Syscall FROM profiles WHERE name = ?`, name).
		Scan(&p.ID, &p.Payload_id, &p.Key, &p.Name, &p.Get, &p.Post, &p.Jitter, &p.Sleep, &p.SleepObf, &p.HeapObf, &p.StackSpoof, &p.Syscall)
	if err != nil {
		return Profile{}, fmt.Errorf("profile not found: %w", err)
	}

	domRows, err := db.conn.Query(`SELECT domain, port, isHttps FROM profile_domains WHERE id = ?`, p.ID)
	if err != nil {
		return Profile{}, err
	}
	defer domRows.Close()
	for domRows.Next() {
		var d ProfileDomain
		if err := domRows.Scan(&d.Domain, &d.Port, &d.IsHttps); err != nil {
			return Profile{}, err
		}
		p.Domains = append(p.Domains, d)
	}

	hdrRows, err := db.conn.Query(`SELECT Key, Value FROM profile_headers WHERE id = ?`, p.ID)
	if err != nil {
		return Profile{}, err
	}
	defer hdrRows.Close()
	for hdrRows.Next() {
		var h ProfileHeader
		if err := hdrRows.Scan(&h.Key, &h.Value); err != nil {
			return Profile{}, err
		}
		p.Headers = append(p.Headers, h)
	}

	return p, nil
}


func (db *DB) DeleteProfile(name string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM profile_domains WHERE id IN (SELECT id FROM profiles WHERE name = ?)", name)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM profile_headers WHERE id IN (SELECT id FROM profiles WHERE name = ?)", name)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM profiles WHERE name = ?", name)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
	
}