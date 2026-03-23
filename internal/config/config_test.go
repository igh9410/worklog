package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	t.Run("keeps regular path", func(t *testing.T) {
		got, err := expandPath("/tmp/worklog")
		if err != nil {
			t.Fatalf("expandPath returned error: %v", err)
		}

		if got != "/tmp/worklog" {
			t.Fatalf("expected /tmp/worklog, got %q", got)
		}
	})

	t.Run("expands home prefix", func(t *testing.T) {
		got, err := expandPath("~/worklog")
		if err != nil {
			t.Fatalf("expandPath returned error: %v", err)
		}

		if got == "~/worklog" {
			t.Fatalf("expected home directory expansion, got %q", got)
		}
	})
}

func TestLoadFromConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	content := `{
  "vault_path": "/tmp/obsidian-vault",
  "daily_notes_dir": "Journal/Daily",
  "timezone": "Asia/Seoul",
  "ignore_repos": [
    "/tmp/obsidian-vault",
    "/tmp/gramnuri"
  ]
}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("WORKLOG_CONFIG_PATH", configPath)
	t.Setenv("WORKLOG_VAULT_PATH", "")
	t.Setenv("WORKLOG_DAILY_NOTES_DIR", "")
	t.Setenv("WORKLOG_TIMEZONE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.VaultPath != "/tmp/obsidian-vault" {
		t.Fatalf("expected vault path from config file, got %q", cfg.VaultPath)
	}
	if cfg.DailyNotesDir != "Journal/Daily" {
		t.Fatalf("expected daily notes dir from config file, got %q", cfg.DailyNotesDir)
	}
	if cfg.Timezone != "Asia/Seoul" {
		t.Fatalf("expected timezone from config file, got %q", cfg.Timezone)
	}
	if len(cfg.IgnoreRepos) != 2 {
		t.Fatalf("expected ignore repos from config file, got %v", cfg.IgnoreRepos)
	}
}

func TestEnvOverridesConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	content := `{
  "vault_path": "/tmp/from-config",
  "daily_notes_dir": "Daily",
  "timezone": "UTC"
}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("WORKLOG_CONFIG_PATH", configPath)
	t.Setenv("WORKLOG_VAULT_PATH", "/tmp/from-env")
	t.Setenv("WORKLOG_DAILY_NOTES_DIR", "Notes/Daily")
	t.Setenv("WORKLOG_TIMEZONE", "Asia/Seoul")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.VaultPath != "/tmp/from-env" {
		t.Fatalf("expected vault path from env, got %q", cfg.VaultPath)
	}
	if cfg.DailyNotesDir != "Notes/Daily" {
		t.Fatalf("expected daily notes dir from env, got %q", cfg.DailyNotesDir)
	}
	if cfg.Timezone != "Asia/Seoul" {
		t.Fatalf("expected timezone from env, got %q", cfg.Timezone)
	}
}

func TestShouldIgnoreRepo(t *testing.T) {
	cfg := &Config{
		VaultPath:   "/home/geonhyuk/Documents/Obsidian Vault",
		IgnoreRepos: []string{"/home/geonhyuk/Documents/CS/Projects/gramnuri"},
	}

	if !cfg.ShouldIgnoreRepo("/home/geonhyuk/Documents/Obsidian Vault") {
		t.Fatalf("expected vault repo to be ignored")
	}

	if !cfg.ShouldIgnoreRepo("/home/geonhyuk/Documents/CS/Projects/gramnuri") {
		t.Fatalf("expected configured repo to be ignored")
	}

	if cfg.ShouldIgnoreRepo("/home/geonhyuk/Documents/CS/Projects/worklog") {
		t.Fatalf("did not expect unrelated repo to be ignored")
	}
}
