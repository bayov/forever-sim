#!/bin/bash
# ab.sh <request-prefix> <preset> <iters> <seed> <variant>...: abtest A = tools/tankbench/req/<prefix>_<preset>.json,
# B = $REQ/<variant>_<preset>.json (gen.py); out $OUT/<prefix>_<preset>__<variant>_s<seed>.txt. Run from the repo root.
BIN=${BIN:-/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/bin}
R=orchestration/tank_handoff/shaman/round5
REQ=${REQ:-/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/r5req}
OUT=${OUT:-$R/out}
mkdir -p "$OUT"
pre=$1; p=$2; it=$3; sd=$4; shift 4
for v in "$@"; do
  "$BIN/abtest" -preset "$p" -iters "$it" -seed "$sd" "tools/tankbench/req/${pre}_$p.json" "$REQ/${pre}/${v}_$p.json" 2>&1 \
    | grep -v -e "Running .* iterations" -e "sims finished" > "$OUT/${pre}_${p}__${v}_s$sd.txt"
done
