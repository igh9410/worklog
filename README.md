# worklog

`worklog` is a small Go CLI that records coding activity into Obsidian daily notes.

The first version is intentionally simple:

- capture the latest Git commit from the current repo
- append a Markdown entry into today's daily note
- record manual notes for work that does not become a commit

## Configuration

You can configure `worklog` either with environment variables or a JSON config file.

Default config path:

```text
~/.config/worklog/config.json
```

Example config:

```json
{
  "vault_path": "/home/your-user/Documents/Obsidian Vault",
  "daily_notes_dir": "Daily",
  "timezone": "Asia/Seoul",
  "ignore_repos": [
    "/home/your-user/Documents/Obsidian Vault"
  ]
}
```

The repository includes a sample at [config.example.json](/home/geonhyuk/Documents/CS/Projects/worklog/config.example.json).

`ignore_repos` is useful when your Obsidian vault is also a git repository, so a repo-local hook does not log vault commits back into the vault.

## Environment

Environment variables override the config file:

```bash
export WORKLOG_VAULT_PATH="$HOME/path/to/obsidian-vault"
export WORKLOG_DAILY_NOTES_DIR="Daily"
export WORKLOG_TIMEZONE="Asia/Seoul"
export WORKLOG_CONFIG_PATH="$HOME/.config/worklog/config.json"
```

`WORKLOG_DAILY_NOTES_DIR`, `WORKLOG_TIMEZONE`, and `WORKLOG_CONFIG_PATH` are optional.

## Commands

Capture the latest commit from the current repository:

```bash
go run ./cmd/worklog capture-commit --repo /path/to/repo
```

Add a manual note:

```bash
go run ./cmd/worklog note add --repo /path/to/repo --message "Investigated OAuth token refresh flow"
```

With your current vault path, the env form would be:

```bash
export WORKLOG_VAULT_PATH="/home/geonhyuk/Documents/Obsidian Vault"
```

## Makefile

For day-to-day setup, use the repository `Makefile`:

```bash
make test
make build
make install
```

For a fresh machine, bootstrap everything in one command:

```bash
make bootstrap \
  WORKLOG_VAULT_PATH="/home/geonhyuk/Documents/Obsidian Vault" \
  WORKLOG_DAILY_NOTES_DIR="." \
  WORKLOG_TIMEZONE="Asia/Seoul"
```

Notes:

- `make install` installs the binary and a `post-commit` hook for `HOOK_REPO` (defaults to the current directory)
- `make install-config` creates `~/.config/worklog/config.json` if it does not exist
- `make overwrite-config` replaces the config file with the provided values
- `WORKLOG_DAILY_NOTES_DIR="."` matches your current vault layout where daily notes live at the vault root
- `make install-hook HOOK_REPO="/path/to/repo"` installs the hook into that repository's `.git/hooks/post-commit`
- `make install-hook` does not change `git config --global core.hooksPath`
- If a repo already has `post-commit`, `make install-hook` moves it to `post-commit.worklog-original` and installs a wrapper that runs both hooks

## Git hook

An example hook is in [scripts/post-commit.example.sh](/home/geonhyuk/Documents/CS/Projects/worklog/scripts/post-commit.example.sh).

Once the CLI is installed, each repository can use a local `post-commit` hook that calls:

```bash
worklog capture-commit --repo "$(git rev-parse --show-toplevel)"
```

To install that hook into another repository from this checkout:

```bash
make install-hook HOOK_REPO="/path/to/repo"
```

If you used an older global hook setup, unset it so repo-local hooks can run again:

```bash
git config --global --unset core.hooksPath
```
