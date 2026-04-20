package kronos

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/z3vxo/kronos/internal/config"
	"gopkg.in/yaml.v3"
)

func SetupConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := setupDirs(home); err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n=== Kronos C2 Setup ===\n")

	username := prompt(reader, "Operator username", "kronos")
	password := prompt(reader, "Operator password", "kronospwd")

	jwtSecret, err := generateSecret(32)
	if err != nil {
		return fmt.Errorf("generating JWT secret: %v", err)
	}

	portStr := prompt(reader, "Teamserver port", "50050")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	certDir := filepath.Join(home, ".kronos", "certs")
	tsCert := filepath.Join(certDir, "server.crt")
	tsKey := filepath.Join(certDir, "server.key")

	fmt.Print("\nGenerating teamserver TLS certs... ")
	if err := generateCert(tsKey, tsCert, "kronos", ""); err != nil {
		return fmt.Errorf("generating teamserver certs: %v", err)
	}
	fmt.Println("done")

	var listenerCert, listenerKey string

	fmt.Print("\nProvide your own SSL cert for listeners? [y/N]: ")
	ownCert, _ := reader.ReadString('\n')

	if yn := strings.TrimSpace(strings.ToLower(ownCert)); yn == "y" || yn == "yes" {
		listenerCert = prompt(reader, "Listener cert path", "")
		listenerKey = prompt(reader, "Listener key path", "")
	} else {
		var domain string
		for domain == "" {
			fmt.Print("Domain for listener cert (required): ")
			line, _ := reader.ReadString('\n')
			domain = strings.TrimSpace(line)
			if domain == "" {
				fmt.Println("  A domain is required to generate a cert.")
			}
		}
		listenerCert = filepath.Join(certDir, "listener.crt")
		listenerKey = filepath.Join(certDir, "listener.key")
		fmt.Printf("Generating listener TLS cert for %s... ", domain)
		if err := generateCert(listenerKey, listenerCert, domain, domain); err != nil {
			return fmt.Errorf("generating listener certs: %v", err)
		}
		fmt.Println("done")
	}

	cfg := config.Config{
		TS: config.TeamServer{
			ListenInterface: "0.0.0.0",
			Port:            port,
			Auth: config.AuthConf{
				Username:          username,
				Password:          password,
				JwtSecret:         jwtSecret,
				TokenHours:        24,
				TokenRefreshHours: 168,
			},
			Cert: tsCert,
			Key:  tsKey,
		},
		Server: config.HttpServer{
			GetEndpoint:  "/ms/download",
			PostEndpoint: "/ms/upload",
			Cert:         listenerCert,
			Key:          listenerKey,
			GetHeaders:   map[string]string{"Server": "Apache"},
			PostHeaders:  map[string]string{"Server": "nginx"},
			NotFoundFile: filepath.Join(home, ".kronos", "404.html"),
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	configFile := filepath.Join(home, ".kronos", "config", "config.yaml")
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return err
	}

	fmt.Printf("\n[*] Setup complete. Config saved to %s\n", configFile)
	fmt.Printfln("[*] View File to update req/resp headers and more")
	fmt.Println("Run './kronos' to start the teamserver.")
	return nil
}

func prompt(r *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := r.ReadString('\n')
	val := strings.TrimSpace(line)
	if val == "" {
		return defaultVal
	}
	return val
}

func generateSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateCert(keyOut, certOut, cn, san string) error {
	args := []string{
		"req", "-x509", "-newkey", "rsa:4096",
		"-days", "365", "-nodes",
		"-keyout", keyOut,
		"-out", certOut,
		"-subj", "/CN=" + cn,
	}
	if san != "" {
		args = append(args, "-addext", "subjectAltName=DNS:"+san)
	}
	var stderr bytes.Buffer
	cmd := exec.Command("openssl", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}
