#!/bin/bash
# opt.sh <preset> <out file> <settings> [rotopt flags...]: rotopt on the six scoring encounters, from the worktree.
W=/tmp/claude-0/-home-user/826fbcec-4fea-56e8-b2bb-2cf6020231d5/scratchpad/warrior
preset=$1; out=$2; settings=$3; shift 3
E=tools/encounters
cd /home/user/forever-sim/.claude/worktrees/agent-a8506f2699ce37e7a
"$W/rotopt" -template warrior_tank -settings "$settings" -tank "$preset" \
  -tank-encounter $E/magmadar.json,$E/golemagg.json,$E/onyxia.json,$E/broodlord.json,$E/vaelastrasz.json,$E/chromaggus.json \
  -iters 600 -confirm-iters 3000 "$@" > "$out" 2>&1
echo "exit $?" >> "$out"
