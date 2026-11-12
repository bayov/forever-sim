#!/bin/bash
# usage: tools/audit/healer_fairness/screen.sh <healbench binary> <role> <out.tsv> <req.json>... [-- extra flags]
# Sims each request (a preset or a hand-edited variant of one) on the 7 boss files in <role> with -carry
# (seed 101, 1,000 iterations unless the extra flags say otherwise) and appends request, role, boss, carry, ehps,
# overheal to <out.tsv>, then prints each request's mean carry. Variants are built with jq from tools/healbench/req/*.json;
# the row label is the file's basename. Run from the repo root.
H=$1; ROLE=$2; OUT=$3; shift 3; REQS=(); while [ $# -gt 0 ] && [ "$1" != "--" ]; do REQS+=("$1"); shift; done; shift; EXTRA="$@"
for q in "${REQS[@]}"; do r=$(basename $q .json)
  for b in onyxia broodlord chromaggus golemagg magmadar ragnaros vaelastrasz; do
    $H -role $ROLE -encounter tools/encounters/$b.json -carry -iters 1000 $EXTRA $q 2>/dev/null | grep "^| $r " |
      awk -F'|' -v r=$r -v role=$ROLE -v b=$b '{for(i=2;i<=NF;i++){gsub(/^ +| +$/,"",$i)}; printf "%s\t%s\t%s\t%s\t%s\t%s\n", r,role,b,$16,$4,$5}' >> $OUT
  done
  awk -F'\t' -v r=$r -v role=$ROLE '$1==r&&$2==role{s+=$4;n++} END{printf "%s %s mean carry %.3f (%d files)\n", r, role, s/n, n}' $OUT
done
