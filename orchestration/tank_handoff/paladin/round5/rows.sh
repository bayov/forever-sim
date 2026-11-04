#!/bin/bash
# rows.sh : the summary.tsv rows, as round 4 made them: tankbench -iters 1000 (default seed), one request and one boss a run
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round5
for p in max_threat balanced max_survival; do for b in magmadar golemagg onyxia broodlord vaelastrasz chromaggus ragnaros; do
  /tmp/claude-0/paladin_r5/tankbench -iters 1000 -encounter tools/encounters/$b.json tools/tankbench/req/paladin_protection_$p.json > $R/rows_raw/paladin_protection_${p}__$b.txt 2>&1; done; done
python3 $R/bench_summary.py
