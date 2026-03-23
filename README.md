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

`ignore_repos` is useful when your Obsidian vault is also a git repository, so the global hook does not log vault commits back into the vault.

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

## Git hook

An example hook is in [scripts/post-commit.example.sh](/home/geonhyuk/Documents/CS/Projects/worklog/scripts/post-commit.example.sh).

Once the CLI is installed, a global Git hook can call:

```bash
worklog capture-commit --repo "$(git rev-parse --show-toplevel)"
```
