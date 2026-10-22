#!/usr/bin/env bash
# Measures character and token savings from ANSI stripping across log fixtures.
#
# Usage:
#   bash devenv/test-responses/build-logs/measure.sh
#
# Requirements: python3 with tiktoken (pip install tiktoken or: nix shell nixpkgs#python3Packages.tiktoken)

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"

python3 << PYEOF
import re, os, warnings
warnings.filterwarnings("ignore")

try:
    import tiktoken
    enc = tiktoken.get_encoding("cl100k_base")
    have_tiktoken = True
except ImportError:
    have_tiktoken = False
    print("WARNING: tiktoken not installed; skipping token counts. Run: pip install tiktoken\n")

ansi = re.compile(r'\x1b(?:[@-Z\\\\-_]|\[[0-9;]*[a-zA-Z])')

fixtures = ["short", "medium", "large", "known"]
dir_ = "$DIR"

rows = []
for name in fixtures:
    path = os.path.join(dir_, f"{name}.txt")
    if not os.path.exists(path):
        continue
    text = open(path, "rb").read().decode("utf-8", errors="replace")
    stripped = ansi.sub("", text)

    chars_before = len(text)
    chars_after  = len(stripped)
    chars_saved  = chars_before - chars_after
    chars_pct    = chars_saved / chars_before * 100 if chars_before else 0

    seqs = len(ansi.findall(text))

    if have_tiktoken:
        tok_before = len(enc.encode(text))
        tok_after  = len(enc.encode(stripped))
        tok_saved  = tok_before - tok_after
        tok_pct    = tok_saved / tok_before * 100 if tok_before else 0
    else:
        tok_before = tok_after = tok_saved = tok_pct = None

    rows.append((name, chars_before, chars_after, chars_saved, chars_pct,
                 seqs, tok_before, tok_after, tok_saved, tok_pct))

# Print results
print(f"{'Fixture':<10} {'Chars before':>14} {'Chars after':>12} {'Saved':>8} {'Saved%':>7}  {'ANSI seqs':>10}")
print("-" * 67)
for r in rows:
    name, cb, ca, cs, cp, seqs, *_ = r
    print(f"{name:<10} {cb:>14,} {ca:>12,} {cs:>8,} {cp:>6.1f}%  {seqs:>10,}")

if have_tiktoken:
    print()
    print(f"{'Fixture':<10} {'Tokens before':>14} {'Tokens after':>13} {'Saved':>8} {'Saved%':>7}")
    print("-" * 57)
    for r in rows:
        name, _, _, _, _, _, tb, ta, ts, tp = r
        print(f"{name:<10} {tb:>14,} {ta:>13,} {ts:>8,} {tp:>6.1f}%")
PYEOF
