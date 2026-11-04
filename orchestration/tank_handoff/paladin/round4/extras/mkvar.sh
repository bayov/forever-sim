#!/bin/bash
# mkvar.sh <variants file> : var/<name>_<preset>.json = var/base_<preset>.json with the variant's consumes (../mkcons.py)
cd /home/user/forever-sim; E=orchestration/tank_handoff/paladin/round4/extras
while read -r n args; do [ -z "$n" ] && continue; for p in max_threat balanced max_survival; do
  python3 $E/../mkcons.py $E/var/${BASE:-base}_$p.json $E/var/${n}_$p.json $args; done; done < "$1"
