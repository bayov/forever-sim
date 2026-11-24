#!/bin/bash
W=/tmp/claude-0/pr; cd $W
it=$1; sd=$2; out=$W/out/grid_i${it}_s$sd.tsv; : > $out
for race in Orc Troll; do for p in balanced max_threat max_survival; do
 for v in racial=0 racial=1 racial=2,racialAt=15 racial=2,racialAt=20 racial=2,racialAt=30 racial=2,racialAt=45 racial=2,racialAt=60 racial=3,racialAt=5 racial=3,racialAt=10 racial=3,racialAt=20 racial=3,racialAt=30 racial=3,racialAt=45; do
  ./opt.sh warrior $p $W/out/g.txt -knobs "$(cat k/warrior_$p),$v" -race $race -eval -iters $it -confirm-iters $it -seed $sd
  echo -e "$race\t$p\t$v\t$(grep -E '^\[0 1 2 3 4 5\]s' $W/out/g.txt | sed 's/.*: //')" >> $out
 done; done; done
