package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type HttpConf struct {
	User     string `yaml:"user"`
	Passwd   string `yaml:"pass"`
	Host     string `yaml:"host"`
	PostPath string `yaml:"postPath"`
	GetPath  string `yaml:"getPath"`
}

type Config struct {
	Http HttpConf `yaml:"http"`
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

	cfg := Config{Http: HttpConf{Host: host, User: user, Passwd: passwd, PostPath: "/api/v2/login", GetPath: "/api/v2/users"}}
	data, err := yaml.Marshal(&cfg)
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
	if err := os.WriteFile(filepath.Join(dir, "client.yaml"), data, 0600); err != nil {
		return err
	}

	profilePath := filepath.Join(dir, "profile.yaml")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return os.WriteFile(profilePath, defaultProfile, 0600)
	}
	return nil
}

var defaultProfile = []byte(`domains:
  - host: "127.0.0.1"
    port: 443
    https: true
  - host: "127.0.0.1"
    port: 80
    https: false

endpoints:
  get: "/ms/download"
  post: "/ms/upload"

headers:
  User-Agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"

sleep:
  interval: 10
  jitter: 20

obfuscation:
  syscall: "none"
  heap: false
  sleep: false
  stack_spoof: false
`)

func LoadCfg() error {
	dir, err := kronosDir()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, "client.yaml"))
	if os.IsNotExist(err) {
		return fmt.Errorf("missing config: run ./client setup")
	}
	if err != nil {
		return err
	}
	Cfg = &Config{}
	return yaml.Unmarshal(data, Cfg)
}

func kronosDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kronos"), nil
}
