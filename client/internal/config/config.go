package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HttpConf struct {
	User   string `json:"user"`
	Passwd string `json:"pass"`
	Host   string `json:"host"`
}

type Config struct {
	Http HttpConf `json:"http"`
}

var Cfg *Config

func Setup() error {
	r := bufio.NewReader(os.Stdin)
	prompt := func(label string) string {
		fmt.Print(label)
		line, _ := r.ReadString('\n')
		return strings.TrimSpace(line)
	}

	host := prompt("Teamserver host (e.g. http://127.0.0.1:50050): ")
	user := prompt("Username: ")
	passwd := prompt("Password: ")

	cfg := Config{Http: HttpConf{Host: host, User: user, Passwd: passwd}}
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return err
	}

	dir, err := kronosDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "client.json"), data, 0600); err != nil {
		return err
	}

	profilePath := filepath.Join(dir, "profile.json")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return os.WriteFile(profilePath, defaultProfile, 0600)
	}
	return nil
}

var defaultProfile = []byte(`{
  "name": "default",
  "domains": [
    { "domain_value": "192.168.1.24", "port": 8080, "is_https": false }
  ],
  "headers": [
    { "key": "User-Agent", "value": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)" },
    { "key": "Accept", "value": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" },
    { "key": "Accept-Language", "value": "en-US,en;q=0.9" }
  ],
  "sleep": 10,
  "jitter": 20,
  "get_endpoint": "/api/v2/users",
  "post_endpoint": "/api/v2/login",
  "sleep_obf": false,
  "heap_obf": false,
  "stack_spoof": false,
  "syscall": 0
}
`)

func LoadCfg() error {
	dir, err := kronosDir()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, "client.json"))
	if os.IsNotExist(err) {
		return fmt.Errorf("missing config: run ./client setup")
	}
	if err != nil {
		return err
	}
	Cfg = &Config{}
	return json.Unmarshal(data, Cfg)
}

func kronosDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kronos"), nil
}
