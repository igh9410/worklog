package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/igh9410/worklog/internal/config"
	"github.com/igh9410/worklog/internal/entry"
)

func AppendDailyEntry(cfg *config.Config, logEntry entry.Entry) (string, error) {
	dailyNotesPath := cfg.DailyNotesPath()
	if err := os.MkdirAll(dailyNotesPath, 0o755); err != nil {
		return "", fmt.Errorf("create daily notes directory: %w", err)
	}

	notePath := filepath.Join(dailyNotesPath, logEntry.CapturedAt.Format("2006-01-02")+".md")
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		if err := os.WriteFile(notePath, []byte(renderDailyHeader(logEntry.CapturedAt)), 0o644); err != nil {
			return "", fmt.Errorf("create daily note: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("stat daily note: %w", err)
	}

	file, err := os.OpenFile(notePath, os.O_APPEND|os.O_RDWR, 0o644)
	if err != nil {
		return "", fmt.Errorf("open daily note: %w", err)
	}
	defer file.Close()

	if err := ensureTrailingNewline(file); err != nil {
		return "", fmt.Errorf("prepare daily note append: %w", err)
	}

	if _, err := file.WriteString(renderEntry(logEntry)); err != nil {
		return "", fmt.Errorf("append daily entry: %w", err)
	}

	return notePath, nil
}

func ensureTrailingNewline(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat daily note: %w", err)
	}
	if info.Size() == 0 {
		return nil
	}

	var tail [1]byte
	if _, err := file.ReadAt(tail[:], info.Size()-1); err != nil {
		return fmt.Errorf("read daily note tail: %w", err)
	}
	if tail[0] == '\n' {
		return nil
	}

	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("append separator newline: %w", err)
	}

	return nil
}

func renderDailyHeader(capturedAt time.Time) string {
	return fmt.Sprintf(`---
type: daily
date: %s
---
# %s

`, capturedAt.Format("2006-01-02"), capturedAt.Format("2006-01-02"))
}

func renderEntry(logEntry entry.Entry) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("## %s\n", logEntry.CapturedAt.Format("15:04")))
	builder.WriteString(fmt.Sprintf("- Type: %s\n", logEntry.Kind))

	if logEntry.RepoName != "" {
		builder.WriteString(fmt.Sprintf("- Repo: `%s`\n", logEntry.RepoName))
	}
	if logEntry.Branch != "" {
		builder.WriteString(fmt.Sprintf("- Branch: `%s`\n", logEntry.Branch))
	}

	switch logEntry.Kind {
	case entry.KindCommit:
		if logEntry.CommitSHA != "" {
			builder.WriteString(fmt.Sprintf("- Commit: `%s`\n", shortSHA(logEntry.CommitSHA)))
		}
		if logEntry.CommitMessage != "" {
			builder.WriteString(fmt.Sprintf("- Message: %s\n", logEntry.CommitMessage))
		}
		if len(logEntry.FilesChanged) > 0 {
			builder.WriteString(fmt.Sprintf("- Files: %s\n", joinInlineCode(logEntry.FilesChanged)))
		}
		if strings.TrimSpace(logEntry.DiffStat) != "" {
			builder.WriteString("- Diffstat:\n")
			builder.WriteString("```text\n")
			builder.WriteString(strings.TrimSpace(logEntry.DiffStat))
			builder.WriteString("\n```\n")
		}
	case entry.KindNote:
		builder.WriteString(fmt.Sprintf("- Note: %s\n", logEntry.Note))
	}

	builder.WriteString("\n")
	return builder.String()
}

func shortSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

func joinInlineCode(values []string) string {
	formatted := make([]string, 0, len(values))
	for _, value := range values {
		formatted = append(formatted, fmt.Sprintf("`%s`", value))
	}
	return strings.Join(formatted, ", ")
}
