#!/bin/bash
# ab.sh <class> <preset> <name> <iters> <seed>: paired abtest, A = the shipped request, B = req/<class>_<name>_<preset>.json
W=/tmp/claude-0/-home-user-forever-sim/16f56354-4609-5851-a77a-5815d0058063/scratchpad/work
c=$1; p=$2; n=$3; it=$4; sd=$5
cd /home/user/forever-sim/.claude/worktrees/agent-a8d0dcc3b585ee4a7
A=tools/tankbench/req/warrior_protection_$p.json; [ $c = druid ] && A=tools/tankbench/req/druid_bear_$p.json
nice $W/abtest -preset $p -iters $it -seed $sd $A $W/req/${c}_${n}_$p.json 2>&1 | grep -v -e "Running .* iterations" -e "sims finished" > $W/ab/${c}_${n}_${p}_i${it}_s$sd.txt
