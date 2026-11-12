#!/bin/bash
# tank_fairness: tankbench every given request on the six scoring bosses.
# usage: bench.sh <outdir> <iters> <seed> <request.json>...   (run from the repo root; TB=tankbench binary)
out=$1; it=$2; seed=$3; shift 3; mkdir -p "$out"
TB=${TB:-/tmp/audit-work/tank_fairness/tankbench}
for r in "$@"; do n=$(basename "$r" .json)
  for b in broodlord chromaggus golemagg magmadar onyxia vaelastrasz; do
    "$TB" -iters "$it" -seed "$seed" -encounter tools/encounters/$b.json "$r" > "$out/${n}__$b.txt" 2>&1
  done; done
