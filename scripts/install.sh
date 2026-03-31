#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin_dir="${YOUTRACK_CLI_BIN_DIR:-$HOME/.local/bin}"
codex_home="${CODEX_HOME:-$HOME/.codex}"
skill_name="youtrack-daily-ops"
skill_source="${repo_root}/skills/${skill_name}"
skill_target="${codex_home}/skills/${skill_name}"
config_example="${repo_root}/.youtrack-cli.yaml.example"
config_target="${HOME}/.youtrack-cli.yaml"

mkdir -p "${bin_dir}" "${codex_home}/skills"

go build -o "${bin_dir}/youtrack-cli" "${repo_root}"
ln -sfn "${skill_source}" "${skill_target}"

if [[ ! -f "${config_target}" && -f "${config_example}" ]]; then
  cp "${config_example}" "${config_target}"
fi

cat <<EOF
Installed:
- CLI: ${bin_dir}/youtrack-cli
- Skill: ${skill_target} -> ${skill_source}

Next:
- Ensure ${bin_dir} is on PATH
- Fill ${config_target} with your YouTrack token if needed
EOF
