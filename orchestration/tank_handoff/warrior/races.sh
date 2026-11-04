#!/bin/bash
# races.sh <preset> <talents> <knobs> <iters> <out>: rotopt -eval for each race on the six scoring encounters.
W=/tmp/claude-0/-home-user/826fbcec-4fea-56e8-b2bb-2cf6020231d5/scratchpad/warrior
preset=$1; talents=$2; knobs=$3; iters=$4; out=$5
: > "$out"
for race in Orc Human Dwarf NightElf Gnome Undead Tauren Troll SkyborneHighOrder SkyborneWindshaper; do
  "$W/opt.sh" "$preset" "$W/race_tmp_$preset.txt" "$W/start.json" -knobs "$knobs" -talents="$talents" -race "$race" -eval -iters "$iters"
  echo "$race $(grep -E '^\[0 1 2 3 4 5\]s' "$W/race_tmp_$preset.txt")" >> "$out"
done
echo done >> "$out"
