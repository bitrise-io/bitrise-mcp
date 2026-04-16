#!/usr/bin/env bash
# Collects before/after response artifacts for verbosity analysis.
#
# Pinned test fixtures (chosen for breadth of field coverage):
#   App:      Bitrise iOS Simple (46b6b9a78a418ee8)
#   Build:    7feee0a4-61f4-4eb6-a019-368cf9ac0d82  (a real finished build)
#   Pipeline: 6ce998c6-cd89-4972-b278-09277ee7e402  (a real finished pipeline)
#
# Usage:
#   BITRISE_TOKEN=<token> bash devenv/test-responses/collect.sh
#
# Requirements: go (via mise), curl, python3, claude CLI

set -euo pipefail

TOKEN="${BITRISE_TOKEN:?BITRISE_TOKEN must be set}"
APP="46b6b9a78a418ee8"
BUILD="7feee0a4-61f4-4eb6-a019-368cf9ac0d82"
PIPELINE="6ce998c6-cd89-4972-b278-09277ee7e402"

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BEFORE="$REPO_ROOT/devenv/test-responses/before"
AFTER="$REPO_ROOT/devenv/test-responses/after"
MCP="$REPO_ROOT/devenv/claude/mcp.json"
CLAUDE="/Users/oliverfalvai/.local/bin/claude"
GO="/Users/oliverfalvai/.local/share/mise/shims/go"
API="https://api.bitrise.io/v0.1"

EXTRACT_PY="$(mktemp /tmp/extract_XXXXXX.py)"
cat > "$EXTRACT_PY" << 'PYEOF'
import sys, json
text = sys.stdin.read().strip()
decoder = json.JSONDecoder()
for i in range(len(text)):
    if text[i] in '{[':
        try:
            obj, _ = decoder.raw_decode(text, i)
            print(json.dumps(obj, indent=2))
            sys.exit(0)
        except Exception:
            pass
print(text, file=sys.stderr)
sys.exit(1)
PYEOF
trap 'kill $SERVER_PID 2>/dev/null || true; rm -f "$EXTRACT_PY"' EXIT

# ── Before: raw Bitrise API responses ────────────────────────────────────────

echo "=== Collecting BEFORE responses ==="

curl_api() {
    local path="$1"; local out="$2"
    # Bitrise API expects the raw PAT as the Authorization value (no Bearer prefix)
    curl -sf -H "Authorization: $TOKEN" "$API$path" \
        | python3 "$EXTRACT_PY" > "$out"
    echo "before/$(basename "$out"): $(wc -c < "$out") bytes"
}

curl_api "/apps/$APP/builds/$BUILD"                  "$BEFORE/get_build.json"
curl_api "/apps/$APP/builds/$BUILD/log/summary"      "$BEFORE/get_build_steps.json"
curl_api "/apps/$APP/builds?limit=3"                 "$BEFORE/list_builds.json"
curl_api "/apps/$APP/pipelines/$PIPELINE"            "$BEFORE/get_pipeline.json"
curl_api "/apps/$APP/pipelines?limit=3"              "$BEFORE/list_pipelines.json"
curl_api "/apps/$APP/branches"                       "$BEFORE/list_branches.json"
curl_api "/apps/$APP/build-workflows"                "$BEFORE/list_build_workflows.json"

# ── After: filtered responses via local MCP server ───────────────────────────

echo ""
echo "=== Starting local MCP server ==="

lsof -ti :8765 | xargs kill -9 2>/dev/null || true
sleep 1

env -u BITRISE_TOKEN ADDR=":8765" "$GO" run "$REPO_ROOT" &
SERVER_PID=$!
sleep 3
echo "Server PID $SERVER_PID"

echo ""
echo "=== Collecting AFTER responses ==="

call_tool() {
    local name="$1"; local prompt="$2"; local out="$3"
    BITRISE_TOKEN="$TOKEN" "$CLAUDE" \
        --strict-mcp-config --mcp-config "$MCP" \
        --dangerously-skip-permissions --tools "" \
        -p "$prompt. Reply with ONLY the raw JSON from the tool result." \
        --output-format text 2>/dev/null \
    | python3 "$EXTRACT_PY" > "$out"
    echo "after/$(basename "$out"): $(wc -c < "$out") bytes"
}

call_tool "get_build" \
    "Use get_build to fetch build $BUILD of app $APP" \
    "$AFTER/get_build.json"

call_tool "get_build_steps" \
    "Use get_build_steps to fetch steps for build $BUILD of app $APP" \
    "$AFTER/get_build_steps.json"

call_tool "list_builds" \
    "Use list_builds to list the 3 most recent builds of app $APP (set limit=3)" \
    "$AFTER/list_builds.json"

call_tool "get_pipeline" \
    "Use get_pipeline to fetch pipeline $PIPELINE of app $APP" \
    "$AFTER/get_pipeline.json"

call_tool "list_pipelines" \
    "Use list_pipelines to list the 3 most recent pipelines of app $APP (set limit=3)" \
    "$AFTER/list_pipelines.json"

call_tool "list_branches" \
    "Use list_branches to list branches of app $APP" \
    "$AFTER/list_branches.json"

call_tool "list_build_workflows" \
    "Use list_build_workflows to list workflows of app $APP" \
    "$AFTER/list_build_workflows.json"

# ── Summary ───────────────────────────────────────────────────────────────────

echo ""
echo "=== Token comparison (cl100k_base) ==="
python3 << PYEOF
import tiktoken, os, warnings
warnings.filterwarnings("ignore")
enc = tiktoken.get_encoding("cl100k_base")
before_dir = "$BEFORE"
after_dir  = "$AFTER"
files = sorted(f for f in os.listdir(before_dir) if f.endswith(".json"))
print(f"{'Tool':<30} {'Before':>7} {'After':>7} {'Saved':>7} {'Saved%':>7}")
print("-" * 62)
for name in files:
    bf = os.path.join(before_dir, name)
    af = os.path.join(after_dir, name)
    if not os.path.exists(af):
        continue
    with open(bf) as f: b = len(enc.encode(f.read()))
    with open(af) as f: a = len(enc.encode(f.read()))
    saved = b - a
    pct = saved / b * 100 if b else 0
    print(f"{name:<30} {b:>7} {a:>7} {saved:>7} {pct:>6.0f}%")
PYEOF
