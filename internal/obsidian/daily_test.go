package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/igh9410/worklog/internal/config"
	"github.com/igh9410/worklog/internal/entry"
)

func TestAppendDailyEntryCreatesAndAppends(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		VaultPath:     tempDir,
		DailyNotesDir: "Daily",
		Timezone:      "Local",
	}

	logEntry := entry.Entry{
		Kind:          entry.KindCommit,
		CapturedAt:    time.Date(2026, 3, 23, 12, 45, 0, 0, time.Local),
		RepoName:      "gramnuri",
		Branch:        "main",
		CommitSHA:     "abcdef123456",
		CommitMessage: "Add worklog support",
		FilesChanged:  []string{"cmd/worklog/main.go", "README.md"},
		DiffStat:      " 2 files changed, 10 insertions(+)",
	}

	notePath, err := AppendDailyEntry(cfg, logEntry)
	if err != nil {
		t.Fatalf("AppendDailyEntry returned error: %v", err)
	}

	wantPath := filepath.Join(tempDir, "Daily", "2026-03-23.md")
	if notePath != wantPath {
		t.Fatalf("expected note path %q, got %q", wantPath, notePath)
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read note: %v", err)
	}

	text := string(content)
	for _, expected := range []string{
		"# 2026-03-23",
		"- Repo: `gramnuri`",
		"- Commit: `abcdef1`",
		"- Message: Add worklog support",
		"```text",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected note to contain %q, got:\n%s", expected, text)
		}
	}
}

func TestAppendDailyEntryAddsNewlineBeforeExistingContent(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		VaultPath:     tempDir,
		DailyNotesDir: "Daily",
		Timezone:      "Local",
	}

	logEntry := entry.Entry{
		Kind:       entry.KindNote,
		CapturedAt: time.Date(2026, 3, 23, 3, 26, 0, 0, time.Local),
		Note:       "Investigated append behavior",
	}

	notePath := filepath.Join(tempDir, "Daily", "2026-03-23.md")
	if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
		t.Fatalf("failed to create daily notes directory: %v", err)
	}
	if err := os.WriteFile(notePath, []byte("# existing"), 0o644); err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	if _, err := AppendDailyEntry(cfg, logEntry); err != nil {
		t.Fatalf("AppendDailyEntry returned error: %v", err)
	}

	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read note: %v", err)
	}

	text := string(content)
	if strings.Contains(text, "# existing## 03:26") {
		t.Fatalf("expected note append to insert a newline before the heading, got:\n%s", text)
	}
	if !strings.Contains(text, "# existing\n## 03:26") {
		t.Fatalf("expected note append to preserve a newline before the heading, got:\n%s", text)
	}
}
