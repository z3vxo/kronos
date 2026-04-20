package kronos

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/z3vxo/kronos/internal/config"
	"github.com/z3vxo/kronos/internal/teamserver"
)

func SetupKronos() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return setupDirs(home)
}

func setupDirs(home string) error {
	basePath := filepath.Join(home, ".kronos")
	for _, dir := range []string{
		filepath.Join(basePath, "config"),
		filepath.Join(basePath, "database"),
		filepath.Join(basePath, "certs"),
		filepath.Join(basePath, "logs"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	for _, f := range []string{
		filepath.Join(basePath, "database", "kronos_db.sql"),
		filepath.Join(basePath, "config", "nyx.log"),
		filepath.Join(basePath, "logs", "kronos.log"),
	} {
		if err := ensureFile(f); err != nil {
			return err
		}
	}
	return nil
}

func validateSetup() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(home, ".kronos", "config", "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("kronos is not configured — run './kronos setup' first")
	}
	return nil
}

func ensureFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		return file.Close()
	}
	return nil
}

func Run() error {
	if err := validateSetup(); err != nil {
		return err
	}
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("failed loading config: %v", err)
	}

	ts, err := teamserver.NewTeamServer()
	if err != nil {
		return fmt.Errorf("failed setting up teamserver: %v", err)
	}

	quit := make(chan os.Signal, 1)
	go func() {
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("Shutting down...")
		ts.Stop()
	}()

	if err := ts.Start(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed starting teamserver: %v", err)
	}
	return nil
}
