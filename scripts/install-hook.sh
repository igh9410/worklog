#!/usr/bin/env bash
set -euo pipefail

repo_arg="${1:-}"
install_bin_path="${2:-}"

if [[ -z "$repo_arg" ]]; then
  printf '%s\n' 'repository path is required' >&2
  exit 1
fi

if [[ -z "$install_bin_path" ]]; then
  printf '%s\n' 'install binary path is required' >&2
  exit 1
fi

repo_root="$(git -C "$repo_arg" rev-parse --show-toplevel 2>/dev/null)" || {
  printf 'HOOK_REPO must point to a Git repository. Current value: %s\n' "$repo_arg" >&2
  exit 1
}

local_hooks_path="$(git -C "$repo_root" config --local --get core.hooksPath 2>/dev/null || true)"
if [[ -n "$local_hooks_path" ]]; then
  if [[ "$local_hooks_path" = /* ]]; then
    hooks_dir="$local_hooks_path"
  else
    hooks_dir="$repo_root/$local_hooks_path"
  fi
else
  git_common_dir="$(git -C "$repo_root" rev-parse --path-format=absolute --git-common-dir)"
  hooks_dir="$git_common_dir/hooks"
fi

hook_path="$hooks_dir/post-commit"
install -d "$hooks_dir"
cat >"$hook_path" <<EOF
#!/usr/bin/env bash
set -euo pipefail

repo_root=\$(git rev-parse --show-toplevel 2>/dev/null || exit 0)
"$install_bin_path" capture-commit --repo "\$repo_root" >/dev/null 2>&1 || true
EOF
chmod 0755 "$hook_path"

global_hooks_path="$(git config --global --get core.hooksPath 2>/dev/null || true)"
if [[ -n "$global_hooks_path" ]]; then
  printf 'Warning: global core.hooksPath is set to %s. Repo-local hooks may not run until it is unset.\n' "$global_hooks_path"
fi

printf 'Installed post-commit hook to %s for repo %s\n' "$hook_path" "$repo_root"
