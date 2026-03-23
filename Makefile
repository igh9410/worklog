APP_NAME := worklog
GO ?= go

LOCAL_BIN_DIR ?= $(CURDIR)/bin
LOCAL_BIN_PATH ?= $(LOCAL_BIN_DIR)/$(APP_NAME)

BIN_DIR ?= $(HOME)/.local/bin
INSTALL_BIN_PATH ?= $(BIN_DIR)/$(APP_NAME)

CONFIG_DIR ?= $(HOME)/.config/worklog
CONFIG_PATH ?= $(CONFIG_DIR)/config.json

HOOKS_DIR ?= $(HOME)/.config/git/hooks
HOOK_PATH ?= $(HOOKS_DIR)/post-commit
CONFIGURE_GIT_HOOKS ?= 1

WORKLOG_VAULT_PATH ?=
WORKLOG_DAILY_NOTES_DIR ?= .
WORKLOG_TIMEZONE ?= Asia/Seoul

.DEFAULT_GOAL := help
.RECIPEPREFIX := >

.PHONY: help fmt test build install bootstrap install-bin install-config overwrite-config install-hook

help:
> @printf '%s\n' \
>   'Targets:' \
>   '  make fmt' \
>   '  make test' \
>   '  make build' \
>   '  make install' \
>   '  make bootstrap WORKLOG_VAULT_PATH="/path/to/Obsidian Vault"' \
>   '  make install-config WORKLOG_VAULT_PATH="/path/to/Obsidian Vault"' \
>   '  make overwrite-config WORKLOG_VAULT_PATH="/path/to/Obsidian Vault"' \
>   '  make install-hook'

fmt:
> gofmt -w cmd/worklog/main.go internal/config/config.go internal/config/config_test.go internal/entry/entry.go internal/gitlog/gitlog.go internal/obsidian/daily.go internal/obsidian/daily_test.go

test:
> $(GO) test ./...

build:
> install -d "$(LOCAL_BIN_DIR)"
> $(GO) build -o "$(LOCAL_BIN_PATH)" ./cmd/worklog
> @printf 'Built %s\n' "$(LOCAL_BIN_PATH)"

install: install-bin install-hook
> @printf '%s\n' 'Installed worklog binary and global post-commit hook.'
> @printf 'If you still need a config file, run: make install-config WORKLOG_VAULT_PATH="%s"\n' '/path/to/Obsidian Vault'

bootstrap: install-bin install-config install-hook
> @printf '%s\n' 'Bootstrapped worklog binary, config, and global post-commit hook.'

install-bin:
> install -d "$(BIN_DIR)"
> $(GO) build -o "$(INSTALL_BIN_PATH)" ./cmd/worklog
> @printf 'Installed binary to %s\n' "$(INSTALL_BIN_PATH)"

install-config:
> @if [ -f "$(CONFIG_PATH)" ]; then \
>   printf 'Config already exists at %s\n' "$(CONFIG_PATH)"; \
>   printf '%s\n' 'Use make overwrite-config if you want to replace it.'; \
>   exit 0; \
> fi
> @if [ -z "$(WORKLOG_VAULT_PATH)" ]; then \
>   printf '%s\n' 'WORKLOG_VAULT_PATH is required.'; \
>   printf 'Example: make install-config WORKLOG_VAULT_PATH="%s" WORKLOG_DAILY_NOTES_DIR="."\n' '/home/your-user/Documents/Obsidian Vault'; \
>   exit 1; \
> fi
> install -d "$(CONFIG_DIR)"
> printf '%s\n' \
>   '{' \
>   '  "vault_path": "$(WORKLOG_VAULT_PATH)",' \
>   '  "daily_notes_dir": "$(WORKLOG_DAILY_NOTES_DIR)",' \
>   '  "timezone": "$(WORKLOG_TIMEZONE)",' \
>   '  "ignore_repos": [' \
>   '    "$(WORKLOG_VAULT_PATH)"' \
>   '  ]' \
>   '}' > "$(CONFIG_PATH)"
> @printf 'Installed config to %s\n' "$(CONFIG_PATH)"

overwrite-config:
> @if [ -z "$(WORKLOG_VAULT_PATH)" ]; then \
>   printf '%s\n' 'WORKLOG_VAULT_PATH is required.'; \
>   printf 'Example: make overwrite-config WORKLOG_VAULT_PATH="%s" WORKLOG_DAILY_NOTES_DIR="."\n' '/home/your-user/Documents/Obsidian Vault'; \
>   exit 1; \
> fi
> install -d "$(CONFIG_DIR)"
> printf '%s\n' \
>   '{' \
>   '  "vault_path": "$(WORKLOG_VAULT_PATH)",' \
>   '  "daily_notes_dir": "$(WORKLOG_DAILY_NOTES_DIR)",' \
>   '  "timezone": "$(WORKLOG_TIMEZONE)",' \
>   '  "ignore_repos": [' \
>   '    "$(WORKLOG_VAULT_PATH)"' \
>   '  ]' \
>   '}' > "$(CONFIG_PATH)"
> @printf 'Wrote config to %s\n' "$(CONFIG_PATH)"

install-hook:
> install -d "$(HOOKS_DIR)"
> printf '%s\n' \
>   '#!/usr/bin/env bash' \
>   'set -euo pipefail' \
>   '' \
>   'repo_root=$$(git rev-parse --show-toplevel 2>/dev/null || exit 0)' \
>   '"$(INSTALL_BIN_PATH)" capture-commit --repo "$$repo_root" >/dev/null 2>&1 || true' > "$(HOOK_PATH)"
> chmod 0755 "$(HOOK_PATH)"
> @if [ "$(CONFIGURE_GIT_HOOKS)" = "1" ]; then \
>   git config --global core.hooksPath "$(HOOKS_DIR)"; \
>   printf 'Configured global core.hooksPath to %s\n' "$(HOOKS_DIR)"; \
> else \
>   printf '%s\n' 'Skipped git config --global core.hooksPath because CONFIGURE_GIT_HOOKS=0.'; \
> fi
> @printf 'Installed hook to %s\n' "$(HOOK_PATH)"
