package installhook

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallHookCreatesRepoLocalHook(t *testing.T) {
	repoDir := initGitRepo(t)
	worklogPath := writeExecutable(t, repoDir, "fake-worklog", `#!/usr/bin/env bash
printf 'worklog:%s\n' "$*" >> "$WORKLOG_TEST_LOG"
`)
	logPath := filepath.Join(repoDir, "hook.log")

	runInstallHook(t, repoDir, worklogPath)

	hookPath := filepath.Join(repoDir, ".git", "hooks", "post-commit")
	runHook(t, repoDir, hookPath, logPath)

	assertFileContains(t, hookPath, "# worklog-managed-hook")
	assertLogLines(t, logPath, []string{
		"worklog:capture-commit --repo " + repoDir,
	})
}

func TestInstallHookSupportsWhitespaceInInstallPath(t *testing.T) {
	repoDir := initGitRepo(t)
	worklogPath := writeExecutable(t, filepath.Join(repoDir, "bin dir"), "fake-worklog", `#!/usr/bin/env bash
printf 'worklog:%s\n' "$*" >> "$WORKLOG_TEST_LOG"
`)
	logPath := filepath.Join(repoDir, "hook.log")

	runInstallHook(t, repoDir, worklogPath)

	hookPath := filepath.Join(repoDir, ".git", "hooks", "post-commit")
	runHook(t, repoDir, hookPath, logPath)

	assertLogLines(t, logPath, []string{
		"worklog:capture-commit --repo " + repoDir,
	})
}

func TestInstallHookChainsExistingPostCommitAndIsIdempotent(t *testing.T) {
	repoDir := initGitRepo(t)
	logPath := filepath.Join(repoDir, "hook.log")
	worklogPath := writeExecutable(t, repoDir, "fake-worklog", `#!/usr/bin/env bash
printf 'worklog:%s\n' "$*" >> "$WORKLOG_TEST_LOG"
`)
	originalHookPath := filepath.Join(repoDir, ".git", "hooks", "post-commit")
	writeExecutableAtPath(t, originalHookPath, `#!/usr/bin/env bash
printf 'existing:%s\n' "$*" >> "$WORKLOG_TEST_LOG"
`)

	runInstallHook(t, repoDir, worklogPath)
	runInstallHook(t, repoDir, worklogPath)

	backupHookPath := filepath.Join(repoDir, ".git", "hooks", "post-commit.worklog-original")
	if _, err := os.Stat(backupHookPath); err != nil {
		t.Fatalf("expected backup hook to exist: %v", err)
	}
	if _, err := os.Stat(backupHookPath + ".worklog-original"); !os.IsNotExist(err) {
		t.Fatalf("expected installer to avoid creating nested backups, got err=%v", err)
	}

	runHook(t, repoDir, originalHookPath, logPath, "arg1", "arg2")

	assertLogLines(t, logPath, []string{
		"existing:arg1 arg2",
		"worklog:capture-commit --repo " + repoDir,
	})
}

func TestInstallHookUsesLocalCoreHooksPath(t *testing.T) {
	repoDir := initGitRepo(t)
	worklogPath := writeExecutable(t, repoDir, "fake-worklog", `#!/usr/bin/env bash
printf 'worklog:%s\n' "$*" >> "$WORKLOG_TEST_LOG"
`)

	runCmd(t, "", "git", "-C", repoDir, "config", "--local", "core.hooksPath", ".githooks")
	runInstallHook(t, repoDir, worklogPath)

	hookPath := filepath.Join(repoDir, ".githooks", "post-commit")
	assertFileContains(t, hookPath, "# worklog-managed-hook")
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()
	runCmd(t, "", "git", "init", repoDir)
	return repoDir
}

func runInstallHook(t *testing.T, repoDir, worklogPath string) {
	t.Helper()

	scriptPath := filepath.Join(repoRoot(t), "scripts", "install-hook.sh")
	runCmd(t, repoRoot(t), "bash", scriptPath, repoDir, worklogPath)
}

func runHook(t *testing.T, repoDir, hookPath, logPath string, args ...string) {
	t.Helper()

	cmd := exec.Command(hookPath, args...)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "WORKLOG_TEST_LOG="+logPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running hook %s failed: %v\n%s", hookPath, err, output)
	}
}

func runCmd(t *testing.T, dir, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %s %s\n%s", name, strings.Join(args, " "), output)
	}
	return string(output)
}

func writeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	writeExecutableAtPath(t, path, content)
	return path
}

func writeExecutableAtPath(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write executable %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("expected %s to contain %q, got:\n%s", path, want, content)
	}
}

func assertLogLines(t *testing.T, path string, want []string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	got := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(got) != len(want) {
		t.Fatalf("unexpected log line count for %s: got %d want %d\n%s", path, len(got), len(want), content)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected log line %d for %s: got %q want %q", i, path, got[i], want[i])
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
