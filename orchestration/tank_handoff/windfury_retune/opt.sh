#!/bin/bash
# opt.sh <class warrior|druid> <preset_underscored> <out> [flags]: rotopt on the six scoring bosses at own weight, request = tools/tankbench/req
c=$1; p=$2; out=$3; shift 3
W=/tmp/claude-0/-home-user-forever-sim/16f56354-4609-5851-a77a-5815d0058063/scratchpad/work
cd /home/user/forever-sim/.claude/worktrees/agent-a8d0dcc3b585ee4a7
E=tools/encounters
if [ $c = warrior ]; then tpl=warrior_tank; req=${REQ:-tools/tankbench/req/warrior_protection_$p.json}; else tpl=druid_tank; req=${REQ:-tools/tankbench/req/druid_bear_$p.json}; fi
nice $W/rotopt -template $tpl -settings "$req" -tank ${p//_/-} \
  -tank-encounter $E/magmadar.json,$E/golemagg.json,$E/onyxia.json,$E/broodlord.json,$E/vaelastrasz.json,$E/chromaggus.json \
  -iters 1000 -confirm-iters 3000 "$@" > "$out" 2>&1
echo "exit $?" >> "$out"
