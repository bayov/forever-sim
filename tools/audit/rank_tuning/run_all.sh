#!/bin/bash
# Usage: AUDIT_TREE=... AUDIT_WORK=... run_all.sh [preset ...]
# With no presets: regenerates allreq.txt from AUDIT_TREE's req dir and sims every base plus every
# jobs.txt sensitivity (about 5 min on 4 cores). With presets: only those bases and their jobs.txt lines.
# Stale 20k inputs for the named presets (or all) are deleted first so each re-reads its req file.
set -euo pipefail
: "${AUDIT_TREE:?}"; : "${AUDIT_WORK:?}"
HERE="$(cd "$(dirname "$0")" && pwd)"; cd "$AUDIT_WORK"
if [ $# -eq 0 ]; then
  ls "$AUDIT_TREE"/tools/dps_report/req/*.json | grep -v '\.meta\.json$' | xargs -n1 basename | sed 's/\.json$//' > allreq.txt
  set -- $(cat allreq.txt); rm -f m2_*.in.json m2_*.out.json
else for n in "$@"; do rm -f "m2_$n.in.json" m2_"$n"_*.out.json; done; fi
for n in "$@"; do "$HERE/run2.sh" "$n" prep; done
( for n in "$@"; do echo "$n base"; grep "^$n " "$HERE/jobs.txt" || true; done ) | xargs -P "$(nproc)" -L1 "$HERE/run2.sh"
