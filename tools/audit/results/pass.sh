#!/bin/bash
# Usage: pass.sh <tag> [iterations=20000]
# One full results-audit pass on AUDIT_TREE: build (skip with SKIP_BUILD=1), sim every DPS report request into
# $AUDIT_WORK/runs/<tag>/, compare with published and with the reference summary, run the drift check and the unread-field
# check, and diff the drift report against the reference. Everything lands in $AUDIT_WORK/runs/<tag>/ (summary.txt,
# compare.txt, drift_report.txt, drift.diff, unread.txt). About 5 minutes at 20k with 6 jobs on 4 cores.
# Env: AUDIT_TREE (default: this repo), AUDIT_WORK (default ${TMPDIR:-/tmp}/audit-results), AUDIT_BIN (default $AUDIT_WORK/bin),
#      AUDIT_REF (default $AUDIT_TREE/orchestration/audit/ref/results), JOBS (default 6).
set -u
D=$(cd "$(dirname "$0")" && pwd)
export AUDIT_TREE=${AUDIT_TREE:-$(cd "$D/../../.." && pwd)}
export AUDIT_WORK=${AUDIT_WORK:-${TMPDIR:-/tmp}/audit-results}
export AUDIT_BIN=${AUDIT_BIN:-$AUDIT_WORK/bin}
AUDIT_REF=${AUDIT_REF:-$AUDIT_TREE/orchestration/audit/ref/results}
tag=$1; it=${2:-20000}; O=$AUDIT_WORK/runs/$tag
mkdir -p "$O"
[ "${SKIP_BUILD:-0}" = 1 ] || bash "$D/build.sh" || exit 1
ls "$AUDIT_TREE"/tools/dps_report/req/*.json | grep -v '\.meta\.json$' \
  | xargs -P "${JOBS:-6}" -I{} bash "$D/runreq.sh" {} "$O" "$it" | sort > "$O/summary.txt"
python3 "$D/compare.py" summary "$O/summary.txt" "$AUDIT_REF/req_summary.txt" > "$O/compare.txt"
python3 "$D/regen.py" > "$O/drift_report.txt"
diff "$AUDIT_REF/drift_report.txt" "$O/drift_report.txt" > "$O/drift.diff"
python3 "$D/unread_fields.py" > "$O/unread.txt"
echo "== $tag at $(git -C "$AUDIT_TREE" rev-parse --short HEAD), $it iterations"
cat "$O/compare.txt"
echo "== drift report vs reference ($(wc -l < "$O/drift.diff") diff lines; empty means unchanged)"
head -40 "$O/drift.diff"
echo "== unread fields (expected: only notes)"
cat "$O/unread.txt"
