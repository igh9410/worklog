package gitlog

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/igh9410/worklog/internal/entry"
)

type RepoContext struct {
	RootPath string
	Name     string
	Branch   string
}

func DetectRepo(path string) (*RepoContext, error) {
	rootPath, err := runGit(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}

	branch, err := runGit(rootPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}

	return &RepoContext{
		RootPath: rootPath,
		Name:     filepath.Base(rootPath),
		Branch:   branch,
	}, nil
}

func CaptureHEAD(path string, capturedAt time.Time) (*entry.Entry, error) {
	repo, err := DetectRepo(path)
	if err != nil {
		return nil, err
	}

	sha, err := runGit(repo.RootPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	message, err := runGit(repo.RootPath, "log", "-1", "--pretty=%s")
	if err != nil {
		return nil, err
	}

	filesOutput, err := runGit(repo.RootPath, "show", "--name-only", "--format=", "HEAD")
	if err != nil {
		return nil, err
	}

	diffStat, err := runGit(repo.RootPath, "show", "--stat", "--format=", "HEAD")
	if err != nil {
		return nil, err
	}

	return &entry.Entry{
		Kind:          entry.KindCommit,
		CapturedAt:    capturedAt,
		RepoName:      repo.Name,
		RepoPath:      repo.RootPath,
		Branch:        repo.Branch,
		CommitSHA:     sha,
		CommitMessage: message,
		FilesChanged:  splitNonEmptyLines(filesOutput),
		DiffStat:      diffStat,
	}, nil
}

func splitNonEmptyLines(value string) []string {
	lines := strings.Split(value, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func runGit(repoPath string, args ...string) (string, error) {
	commandArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", commandArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}

	return strings.TrimSpace(string(output)), nil
}
