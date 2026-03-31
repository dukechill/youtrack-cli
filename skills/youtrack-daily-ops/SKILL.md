---
name: youtrack-daily-ops
description: YouTrack daily ticket triage and end-of-day update workflow for Duke Chiu. Use when reviewing assigned tickets, updating State, Boards, Assignee, Priority, or logging concise progress/work items.
---

# YouTrack Daily Ops

Use this skill for recurring YouTrack ticket management work.

## Scope
- Prefer `youtrack-cli` from `PATH`.
- If it is not on `PATH`, look for a repo-local binary next to this skill's repository.
- Reuse config from `~/.youtrack-cli.yaml`.
- Use these commands before falling back to MCP or manual action lists:
  - `youtrack-cli list`
  - `youtrack-cli work check`
  - `youtrack-cli work add <issue-id> <minutes> <description>`
  - `youtrack-cli issue comment <issue-id> <text>`
  - `youtrack-cli issue set-estimation <issue-id> <minutes>`
  - `youtrack-cli issue set-state <issue-id> <state>`
  - `youtrack-cli issue daily-sync <issue-id> --minutes <n> --state "<state>" --comment "<text>"`
- Primary board: `https://eslite.youtrack.cloud/agiles/121-114/current?query=%23dukechiu%20`
- Primary personal query: `#dukechiu`
- Default fields to inspect or update: `State`, `Boards`, `Assignee`, `Priority`
- Default sprint source: trust `youtrack-cli list` to resolve the current sprint from YouTrack board data unless the user explicitly names a sprint.

## Default workflow

### Morning triage
- Find tickets from the primary board or query.
- When the user says `current sprint`, prefer the CLI-resolved sprint instead of guessing from the largest sprint number seen in old output.
- Group them into `needs action today`, `blocked`, and `waiting`.
- Recommend at most 3 tickets to actively push today.
- Avoid changing fields during triage unless the current state is clearly stale.

### End-of-day sync
- Focus only on tickets actually touched today.
- For each touched ticket:
  - log work if the user wants time tracking
  - set `Estimation` if the user refers to fixed time, rough estimate, or expected effort
  - add one concise progress update
  - update `State` only if the actual execution status changed
- Keep updates short and factual.

## PR and branch rules
- If a branch name includes a ticket key like `CT-4519`, infer that ticket as the candidate issue.
- If there is an open PR for the ticket branch, set `State` to `Review` unless the user says otherwise.
- Prefer checking the repo branch name and PR status before asking the user which ticket to update.

## Time rules
- Treat logged work (`Spent time`) and fixed time (`Estimation`) as different fields.
- `Spent time` is actual time already used.
- `Estimation` is the rough expected effort.
- If the user says `固定工時` or describes a coarse estimate, update `Estimation`, not just `Spent time`.
- If the user gives one number for both actual and rough estimate, it is acceptable to set both fields to the same minutes when that matches the user's intent.

## Progress note format
- Use one line when possible: `Today: <done>. Next: <next step>. Risk: <blocker or none>.`
- If there is no blocker, use `Risk: none`.

## Field guardrails
- Do not update fields outside `State`, `Boards`, `Assignee`, `Priority`, and `Estimation` unless the user asks.
- Do not reassign tickets or move boards in bulk without explicit user approval.
- Prefer changing `State` over writing long comments when the status already communicates the situation.
- If several tickets need the same update, propose a batch operation before applying it.

## When MCP is unavailable
- If the local CLI is available, use it instead of preparing manual steps.
- Do not pretend changes were applied.
- Instead, prepare a compact action list for the user:
  - target issue IDs
  - intended field changes
  - intended worklog durations
  - intended progress note text

## Response style
- Be concise.
- Default to a short table or flat list for ticket summaries.
- Call out stale or contradictory tickets directly, for example:
  - `State=In Progress` but no recent progress note
  - `Priority` is high but not in today's top 3
