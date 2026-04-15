package database

import (
	"fmt"
	"os"
)

func GetDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/database/%s", homeDir, "nyx_db.sql"), nil
}

func CheckAndSetupDB() error {
	dbPath, err := GetDBPath()
	if err != nil {
		return err
	}

	return nil
}
