#!/bin/bash
W=/tmp/claude-0/pr; cd $W; out=$W/out/stone.tsv; : > $out
for p in balanced max_threat max_survival; do for at in 0 5 10 20 30 60; do
  ./opt2.sh warrior $p $W/out/s.txt -knobs "$(cat k/warrior_$p),stoneformAt=$at" -race Dwarf -eval -iters 10000 -confirm-iters 10000 -seed 1
  printf "%s\t%s\t%s\n" $p $at "$(grep -E '^\[0 1 2 3 4 5\]s' $W/out/s.txt | sed 's/.*: //')" >> $out
done; done
