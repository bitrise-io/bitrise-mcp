# MCP Tool Verbosity Report

Token counts use the `cl100k_base` encoding (tiktoken), matching the tokeniser used by the Claude model family.

## Test fixtures

| Constant | Value |
|---|---|
| App | `46b6b9a78a418ee8` (Bitrise iOS Simple) |
| Build | `7feee0a4-61f4-4eb6-a019-368cf9ac0d82` |
| Pipeline | `6ce998c6-cd89-4972-b278-09277ee7e402` |

Responses collected via `devenv/test-responses/collect.sh`. Before = raw Bitrise API response. After = filtered response from local MCP server (`verbose=false`).

---

## Summary

| Tool | Before (tokens) | After (tokens) | Saved | Saved% |
|---|---:|---:|---:|---:|
| `get_build` | 581 | 331 | 250 | 43% |
| `get_build_steps` | 1,786 | 831 | 955 | 53% |
| `list_builds` | 1,179 | 880 | 299 | 25% |
| `get_pipeline` | 4,247 | 3,734 | 513 | 12% |
| `list_pipelines` | 2,081 | 724 | 1,357 | 65% |
| `list_branches` | 72,435 | 960 | 71,475 | 99% |
| `list_build_workflows` | 544 | 544 | 0 | 0% |

`list_build_workflows` returns a flat array of workflow name strings — nothing to filter.

---

## Per-tool breakdown

### `get_build` — 43% reduction (581 → 331 tokens)

| Field | Treatment | Reason |
|---|---|---|
| `environment_prepare_finished_at` | **Always removed** | Internal pipeline scheduling timestamp |
| `is_processed` | **Always removed** | Internal delivery state |
| `is_status_sent` | **Always removed** | Internal delivery state |
| `log_format` | **Always removed** | Internal log format flag |
| `pull_request_id` when 0 | **Always removed** | Zero means non-PR build; omitting avoids forcing callers to check |
| `original_build_params` | Behind `verbose` | Duplicates `branch`, `commit_hash`, `commit_message`; adds internal PR head/merge branch refs already captured by `pull_request_id` |
| `credit_cost` | Behind `verbose` | Billing metadata |
| `commit_view_url` | Behind `verbose` | Reconstructable from repo URL + `commit_hash` |

---

### `get_build_steps` — 53% reduction (1,786 → 831 tokens)

| Field | Treatment | Reason |
|---|---|---|
| `app_id` | **Always removed** | Caller's own request parameter |
| `build_id` | **Always removed** | Caller's own request parameter |
| `agent_info` (full object) | **Always removed** | Runner infrastructure details already present on the build object |
| Per-step: `collection` | **Always removed** | Always the same steplib URL |
| Per-step: `support_url` | **Always removed** | Always the same steplib URL |
| Per-step: `release_notes` | **Always removed** | Registry changelog, not relevant to this build's execution |
| Per-step: `source_code_url` | Kept | Useful for debugging step behaviour |
| Per-step: `latest_version` | Kept | Useful for spotting outdated step versions |
| `cli_info` | Behind `verbose` | Dump of internal boolean CLI flags (`ci_mode`, `pr_mode`, etc.) that duplicate build-level fields |
| `has_build_environment_setup_logs` | Behind `verbose` | Internal setup log flag |
| `is_log_archived` | Behind `verbose` | Internal archival flag |

---

### `list_builds` — 25% reduction (1,179 → 880 tokens)

The test fixture is a PR build; reduction is smaller for push/scheduled builds where `original_build_params` is minimal.

| Field | Treatment | Reason |
|---|---|---|
| `credit_cost` | **Always removed** | Billing metadata repeated on every row |
| `commit_view_url` | **Always removed** | Reconstructable from repo URL + `commit_hash` |
| `environment_prepare_finished_at` | **Always removed** | Internal scheduling timestamp |
| `is_processed` | **Always removed** | Internal delivery state |
| `is_status_sent` | **Always removed** | Internal delivery state |
| `log_format` | **Always removed** | Internal log format flag |
| `pull_request_id` when 0 | **Always removed** | Omitted when zero |
| `original_build_params` | Behind `verbose` | Duplicates top-level trigger fields; adds internal PR branch refs |
| `repository` (app-scoped request) | Behind `verbose` | Entire object removed — same app metadata repeated on every row |
| `repository` (global request) | Behind `verbose` | Collapsed to `{slug, title, repo_owner, repo_slug}` for app identification only |

---

### `get_pipeline` — 12% reduction (4,247 → 3,734 tokens)

Lower saving because the response is dominated by the `workflows` array (one entry per workflow), which is not filtered.

| Field | Treatment | Reason |
|---|---|---|
| `app` (`{slug}`) | **Always removed** | Caller's own request parameter |
| `number_in_app_scope` | **Always removed** | Internal sequence counter |
| `put_on_hold_at` | **Always removed** | Almost always null; covered by `status` |
| Per-workflow: `startFailureReason` when empty | **Always removed** | Present on every workflow even when no failure occurred; kept only when non-empty |
| `trigger_params` | Behind `verbose` | Overlaps with top-level trigger fields; `environments` array can carry large agent configs |
| `attempts` | Behind `verbose` | Retry history; `current_attempt_id` at top level is sufficient |
| `credit_cost` | Behind `verbose` | Billing metadata |

---

### `list_pipelines` — 65% reduction (2,081 → 724 tokens)

Large saving driven by `trigger_params` removal. For AI-agent pipeline builds the `environments` array inside `trigger_params` contains ~80-entry domain allowlists that bloat every list item significantly. Note that `trigger_params` is unconditionally removed in the list view but is available behind `verbose` in `get_pipeline`.

| Field | Treatment | Reason |
|---|---|---|
| `credit_cost` | **Always removed** | Billing metadata repeated on every row |
| `is_processed` | **Always removed** | Internal delivery state |
| `pull_request_id` when 0 | **Always removed** | Omitted when zero |
| `trigger_params` | **Always removed** | Full trigger input copy including large `environments` array; top-level `branch`, `commit_hash`, `commit_message` already provide the essential context |

---

### `list_branches` — 99% reduction (72,435 → 960 tokens)

The Bitrise API returns all branches in a single response with no server-side pagination. The test app has 2,998 branches.

| Change | Treatment | Reason |
|---|---|---|
| Branch list truncated to 50 by default | **Always applied** | Branches are pre-sorted by most-recently-built; the default limit returns the most relevant ones |
| `meta.returned` / `meta.total` added | **Always added** | Lets the caller know how many branches exist and how many were returned |
| `limit` parameter | Configurable | Caller can raise or lower the cap |

---

### `list_build_workflows` — 0% reduction (544 → 544 tokens)

Returns a flat array of workflow name strings. Nothing to filter.
