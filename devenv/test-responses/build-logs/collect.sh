#!/usr/bin/env bash
# Collects raw build log fixtures (short, medium, large) for ANSI-stripping analysis.
#
# Pinned app: Bitrise iOS Simple (46b6b9a78a418ee8)
# Fetches the 10 most recent builds, downloads their raw logs, then picks
# the smallest, a medium-sized one, and the largest as the three fixtures.
#
# Usage:
#   BITRISE_TOKEN=<token> bash devenv/test-responses/build-logs/collect.sh
#
# Requirements: curl, python3 (with tiktoken: pip install tiktoken)

set -euo pipefail

TOKEN="${BITRISE_TOKEN:?BITRISE_TOKEN must be set}"
APP="46b6b9a78a418ee8"
API="https://api.bitrise.io/v0.1"
OUT="$(cd "$(dirname "$0")" && pwd)"

echo "=== Listing recent builds ==="

# status=1 (success) and status=2 (failed) both have archived logs
builds_json=$(curl -sf -H "Authorization: $TOKEN" "$API/apps/$APP/builds?limit=20")
build_slugs=$(echo "$builds_json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for b in data.get('data', []):
    if b.get('status', 0) != 0:  # 0 = still running
        print(b['slug'])
")

echo "Found builds:"
echo "$build_slugs"

# Download log for each build, save to temp files with size info
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

declare -A log_sizes
declare -A log_files

for slug in $build_slugs; do
    echo -n "  Fetching log for $slug ... "
    log_meta=$(curl -sf -H "Authorization: $TOKEN" "$API/apps/$APP/builds/$slug/log" || true)
    url=$(echo "$log_meta" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get('expiring_raw_log_url', ''))
" 2>/dev/null || true)

    if [ -z "$url" ]; then
        echo "no URL (build still running?), skipping"
        continue
    fi

    tmpfile="$tmpdir/$slug.txt"
    # Logs are plain text (not JSON) at the expiring URL
    if curl -sf --max-time 30 "$url" -o "$tmpfile" 2>/dev/null; then
        size=$(wc -c < "$tmpfile")
        lines=$(wc -l < "$tmpfile")
        echo "${size} bytes, ${lines} lines"
        log_sizes[$slug]=$size
        log_files[$slug]=$tmpfile
    else
        echo "download failed, skipping"
    fi
done

if [ ${#log_files[@]} -eq 0 ]; then
    echo "ERROR: no logs downloaded"
    exit 1
fi

echo ""
echo "=== Selecting short / medium / large fixtures ==="

# Sort by size and pick three evenly-spaced fixtures
python3 << PYEOF
import os, shutil

files = {
$(for slug in "${!log_files[@]}"; do
    echo "    '$slug': ('${log_files[$slug]}', ${log_sizes[$slug]}),"
done)
}

sorted = sorted(files.items(), key=lambda x: x[1][1])
n = len(sorted)
out = "$OUT"

picks = {}
if n == 1:
    picks = {"medium": sorted[0]}
elif n == 2:
    picks = {"short": sorted[0], "large": sorted[-1]}
else:
    mid_idx = n // 2
    picks = {"short": sorted[0], "medium": sorted[mid_idx], "large": sorted[-1]}

for label, (slug, (path, size)) in picks.items():
    dest = os.path.join(out, f"{label}.txt")
    shutil.copy(path, dest)
    lines = open(path).read().count('\n')
    print(f"  {label}: {slug}  ({size:,} bytes, {lines:,} lines)")
PYEOF

echo ""
echo "Fixtures saved to $OUT/"
ls -lh "$OUT"/*.txt
