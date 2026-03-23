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
