#!/usr/bin/env bash
set -euo pipefail

worklog capture-commit --repo "$(git rev-parse --show-toplevel)"
