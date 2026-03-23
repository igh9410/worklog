package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	VaultPath     string
	DailyNotesDir string
	Timezone      string
	IgnoreRepos   []string
}

type fileConfig struct {
	VaultPath     string   `json:"vault_path"`
	DailyNotesDir string   `json:"daily_notes_dir"`
	Timezone      string   `json:"timezone"`
	IgnoreRepos   []string `json:"ignore_repos"`
}

func Load() (*Config, error) {
	configPath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	fileCfg, err := loadFileConfig(configPath)
	if err != nil {
		return nil, err
	}

	vaultPath := getEnvOrDefault("WORKLOG_VAULT_PATH", fileCfg.VaultPath)
	expandedVaultPath, err := expandPath(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("expand WORKLOG_VAULT_PATH: %w", err)
	}

	if strings.TrimSpace(expandedVaultPath) == "" {
		return nil, fmt.Errorf("WORKLOG_VAULT_PATH is required or set vault_path in %s", configPath)
	}

	cfg := &Config{
		VaultPath:     expandedVaultPath,
		DailyNotesDir: getEnvOrDefault("WORKLOG_DAILY_NOTES_DIR", firstNonEmpty(fileCfg.DailyNotesDir, "Daily")),
		Timezone:      getEnvOrDefault("WORKLOG_TIMEZONE", firstNonEmpty(fileCfg.Timezone, "Local")),
		IgnoreRepos:   normalizePaths(fileCfg.IgnoreRepos),
	}

	return cfg, nil
}

func (c *Config) DailyNotesPath() string {
	return filepath.Join(c.VaultPath, c.DailyNotesDir)
}

func (c *Config) ShouldIgnoreRepo(repoPath string) bool {
	cleanRepoPath := filepath.Clean(repoPath)
	if cleanRepoPath == filepath.Clean(c.VaultPath) {
		return true
	}

	for _, ignoredPath := range c.IgnoreRepos {
		if cleanRepoPath == filepath.Clean(ignoredPath) {
			return true
		}
	}

	return false
}

func (c *Config) Location() (*time.Location, error) {
	if c.Timezone == "" || strings.EqualFold(c.Timezone, "Local") {
		return time.Local, nil
	}

	location, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", c.Timezone, err)
	}

	return location, nil
}

func (c *Config) Now() (time.Time, error) {
	location, err := c.Location()
	if err != nil {
		return time.Time{}, err
	}

	return time.Now().In(location), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func configFilePath() (string, error) {
	if explicitPath := strings.TrimSpace(os.Getenv("WORKLOG_CONFIG_PATH")); explicitPath != "" {
		expandedPath, err := expandPath(explicitPath)
		if err != nil {
			return "", fmt.Errorf("expand WORKLOG_CONFIG_PATH: %w", err)
		}
		return expandedPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "worklog", "config.json"), nil
}

func loadFileConfig(path string) (*fileConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileConfig{}, nil
		}
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	var cfg fileConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	return &cfg, nil
}

func normalizePaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}

		expandedPath, err := expandPath(trimmed)
		if err != nil {
			result = append(result, trimmed)
			continue
		}

		result = append(result, expandedPath)
	}

	return result
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/")), nil
	}

	return path, nil
}
