#!/bin/bash
# races.sh <preset> <knobs> <iters> <out> <request>: rotopt -eval for each race on the six scoring encounters.
cd /home/user/forever-sim; R=orchestration/tank_handoff/warrior/round3
preset=$1; knobs=$2; iters=$3; out=$4; req=$5
: > "$out"
for race in Orc Human Dwarf NightElf Gnome Undead Tauren Troll SkyborneHighOrder SkyborneWindshaper; do
  $R/opt.sh "$preset" "$out.tmp" "$req" -knobs "$knobs" -race "$race" -eval -iters "$iters" -confirm-iters "$iters"
  echo "$race $(grep -E '^\[0 1 2 3 4 5\]s' "$out.tmp")" >> "$out"
done
rm -f "$out.tmp"; echo done >> "$out"
