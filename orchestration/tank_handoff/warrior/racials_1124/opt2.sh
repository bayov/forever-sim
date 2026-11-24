#!/bin/bash
# opt.sh <class warrior|druid> <preset_underscored> <out> [flags]: rotopt on the six scoring bosses at own weight, request = tools/tankbench/req
c=$1; p=$2; out=$3; shift 3
W=/tmp/claude-0/pr
cd /home/user/forever-sim
E=tools/encounters
if [ $c = warrior ]; then tpl=warrior_tank; req=${REQ:-tools/tankbench/req/warrior_protection_$p.json}; else tpl=druid_tank; req=${REQ:-tools/tankbench/req/druid_bear_$p.json}; fi
nice $W/new/rotopt -template $tpl -settings "$req" -tank ${p//_/-} \
  -tank-encounter $E/magmadar.json,$E/golemagg.json,$E/onyxia.json,$E/broodlord.json,$E/vaelastrasz.json,$E/chromaggus.json \
  -iters 1000 -confirm-iters 3000 "$@" > "$out" 2>&1
echo "exit $?" >> "$out"
