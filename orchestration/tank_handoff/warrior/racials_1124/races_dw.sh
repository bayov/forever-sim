#!/bin/bash
W=/tmp/claude-0/pr; cd $W
it=$1; sd=$2; out=$W/out/races_dwarf.tsv; : > $out
for p in balanced max_threat max_survival; do for race in Dwarf; do
  ./opt2.sh warrior $p $W/out/r_$sd.txt -knobs "$(cat k/warrior_$p)" -race $race -eval -iters $it -confirm-iters $it -seed $sd
  echo -e "$p\t$race\t$(grep -E '^\[0 1 2 3 4 5\]s' $W/out/r_$sd.txt | sed 's/.*: //')" >> $out
done; done
