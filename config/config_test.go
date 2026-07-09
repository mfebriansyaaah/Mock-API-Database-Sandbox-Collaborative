package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Save and clear the env keys we test so we exercise the defaults.
	keys := []string{EnvPort, EnvProjectID, EnvLogChannelBuffer, EnvLogNumWorkers, EnvLogCleanupInterval, EnvLogMaxLogsPerProject}
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		saved[k] = os.Getenv(k)
		_ = os.Unsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	})

	c := Load()

	if c.Port != defaultPort {
		t.Errorf("Port: want %q, got %q", defaultPort, c.Port)
	}
	if c.ProjectID != "mockapi-sandbox-dev" {
		t.Errorf("ProjectID: want fallback, got %q", c.ProjectID)
	}
	if c.Logger.ChannelBuffer != defaultLogChannelBuffer {
		t.Errorf("ChannelBuffer: want %d, got %d", defaultLogChannelBuffer, c.Logger.ChannelBuffer)
	}
	if c.Logger.NumWorkers != defaultLogNumWorkers {
		t.Errorf("NumWorkers: want %d, got %d", defaultLogNumWorkers, c.Logger.NumWorkers)
	}
	if c.Logger.CleanupInterval != defaultLogCleanupInterval {
		t.Errorf("CleanupInterval: want %s, got %s", defaultLogCleanupInterval, c.Logger.CleanupInterval)
	}
	if c.Logger.MaxLogsPerProject != defaultLogMaxLogsPerProject {
		t.Errorf("MaxLogsPerProject: want %d, got %d", defaultLogMaxLogsPerProject, c.Logger.MaxLogsPerProject)
	}
}

func TestLoad_Override(t *testing.T) {
	t.Setenv(EnvPort, "9999")
	t.Setenv(EnvProjectID, "test-project")
	t.Setenv(EnvLogChannelBuffer, "42")
	t.Setenv(EnvLogNumWorkers, "7")
	t.Setenv(EnvLogCleanupInterval, "30s")
	t.Setenv(EnvLogMaxLogsPerProject, "250")

	c := Load()
	if c.Port != "9999" {
		t.Errorf("Port override: got %q", c.Port)
	}
	if c.ProjectID != "test-project" {
		t.Errorf("ProjectID override: got %q", c.ProjectID)
	}
	if c.Logger.ChannelBuffer != 42 {
		t.Errorf("ChannelBuffer override: got %d", c.Logger.ChannelBuffer)
	}
	if c.Logger.NumWorkers != 7 {
		t.Errorf("NumWorkers override: got %d", c.Logger.NumWorkers)
	}
	if c.Logger.CleanupInterval != 30*time.Second {
		t.Errorf("CleanupInterval override: got %s", c.Logger.CleanupInterval)
	}
	if c.Logger.MaxLogsPerProject != 250 {
		t.Errorf("MaxLogsPerProject override: got %d", c.Logger.MaxLogsPerProject)
	}
}

func TestLoad_InvalidFallsBackToDefault(t *testing.T) {
	t.Setenv(EnvLogChannelBuffer, "not-a-number")
	c := Load()
	if c.Logger.ChannelBuffer != defaultLogChannelBuffer {
		t.Errorf("invalid int should fall back to default, got %d", c.Logger.ChannelBuffer)
	}
}
