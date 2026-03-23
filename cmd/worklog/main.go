package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/igh9410/worklog/internal/config"
	"github.com/igh9410/worklog/internal/entry"
	"github.com/igh9410/worklog/internal/gitlog"
	"github.com/igh9410/worklog/internal/obsidian"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "worklog: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "capture-commit":
		return runCaptureCommit(args[1:])
	case "note":
		return runNote(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runCaptureCommit(args []string) error {
	flags := flag.NewFlagSet("capture-commit", flag.ContinueOnError)
	repoPath := flags.String("repo", ".", "path to the git repository")
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	now, err := cfg.Now()
	if err != nil {
		return err
	}

	logEntry, err := gitlog.CaptureHEAD(*repoPath, now)
	if err != nil {
		return err
	}

	if cfg.ShouldIgnoreRepo(logEntry.RepoPath) {
		fmt.Printf("skipped commit capture for ignored repo %s\n", logEntry.RepoName)
		return nil
	}

	notePath, err := obsidian.AppendDailyEntry(cfg, *logEntry)
	if err != nil {
		return err
	}

	fmt.Printf("captured commit in %s\n", notePath)
	return nil
}

func runNote(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing note subcommand")
	}

	switch args[0] {
	case "add":
		return runNoteAdd(args[1:])
	default:
		return fmt.Errorf("unknown note subcommand %q", args[0])
	}
}

func runNoteAdd(args []string) error {
	flags := flag.NewFlagSet("note add", flag.ContinueOnError)
	message := flags.String("message", "", "manual work note")
	repoPath := flags.String("repo", ".", "path to the related git repository")
	flags.SetOutput(os.Stderr)
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *message == "" && flags.NArg() > 0 {
		*message = strings.Join(flags.Args(), " ")
	}
	if strings.TrimSpace(*message) == "" {
		return fmt.Errorf("note message is required")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	now, err := cfg.Now()
	if err != nil {
		return err
	}

	repo, err := gitlog.DetectRepo(*repoPath)
	if err != nil {
		repo = &gitlog.RepoContext{}
	}
	if repo.RootPath != "" && cfg.ShouldIgnoreRepo(repo.RootPath) {
		fmt.Printf("skipped note capture for ignored repo %s\n", repo.Name)
		return nil
	}

	logEntry := entry.Entry{
		Kind:       entry.KindNote,
		CapturedAt: now,
		RepoName:   repo.Name,
		RepoPath:   repo.RootPath,
		Branch:     repo.Branch,
		Note:       strings.TrimSpace(*message),
	}

	notePath, err := obsidian.AppendDailyEntry(cfg, logEntry)
	if err != nil {
		return err
	}

	fmt.Printf("captured note in %s\n", notePath)
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stdout, `worklog records coding activity into Obsidian daily notes.

Usage:
  worklog capture-commit [--repo PATH]
  worklog note add --message "Investigated login issue" [--repo PATH]

Environment:
  WORKLOG_VAULT_PATH       absolute path to your Obsidian vault
  WORKLOG_DAILY_NOTES_DIR  daily note directory inside the vault (default: Daily)
  WORKLOG_TIMEZONE         IANA timezone name or Local (default: Local)

Example:
  WORKLOG_VAULT_PATH="$HOME/Documents/obsidian-vault" worklog capture-commit
`)
}
