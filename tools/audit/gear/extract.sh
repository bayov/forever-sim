#!/usr/bin/env bash
# Extracts every Level 60 synthetic set from a git ref into a flat dir the checkers read.
# Usage: AUDIT_TREE=<checkout> tools/audit/gear/extract.sh <ref, e.g. origin/master> <outdir, e.g. $AUDIT_WORK/s33>
# Writes <class>.go (forever_synthetic_items_<class>.go), shaman_enh.go (the base forever_synthetic_items.go)
# and master--<path with / as _>.gear.json for every ui/**/level60*.gear.json. lin2.py takes the archetype
# from the part of the file name after '--', so keep that prefix.
set -euo pipefail
ref=$1; out=$2; cd "${AUDIT_TREE:-.}"
rm -rf "$out"; mkdir -p "$out"
for c in warrior paladin hunter rogue priest mage warlock druid shaman; do
  git show "$ref:tools/database/forever_synthetic_items_$c.go" > "$out/$c.go" 2>/dev/null || rm -f "$out/$c.go"
done
git show "$ref:tools/database/forever_synthetic_items.go" > "$out/shaman_enh.go"
for f in $(git ls-tree -r --name-only "$ref" ui | grep -E 'level60.*\.gear\.json$'); do
  git show "$ref:$f" > "$out/master--$(echo "$f" | tr / _)"
done
ls "$out" | wc -l
