#!/bin/bash
# Sims every build whose request exists and has no output yet; loops until STOP exists and nothing is left.
# Works in $DASH_DIR (default: this directory). Start up to 4 of these side by side on 4 CPUs.
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "${DASH_DIR:-$HERE}" || exit 1
mkdir -p out
while true; do
  did=0
  for m in req/*.meta.json; do
    [ -e "$m" ] || continue
    id=$(basename "$m" .meta.json)
    [ -e "out/$id.json" ] && continue
    [ -e "req/$id.json" ] || continue
    # skip files modified in the last 20 s (a helper may still be writing)
    [ $(( $(date +%s) - $(stat -c %Y "$m") )) -lt 20 ] && continue
    python3 "$HERE/run_build.py" "$id" 2>&1 | tail -3
    did=1
  done
  [ $did -eq 0 ] && [ -e STOP ] && break
  [ $did -eq 0 ] && sleep 30
done
echo RUNNER DONE
