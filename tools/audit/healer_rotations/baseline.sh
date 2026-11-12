#!/bin/bash
# usage: tools/audit/healer_rotations/baseline.sh <healbench binary> <out.tsv> [iters] [extra healbench flags...]
# Sims every healer request on every boss file in its role (-carry) and writes the summary.tsv format
# of orchestration/audit/ref/healer_bench/. Run from the repo root.
H=$1; OUT=$2; IT=${3:-1000}; shift 3 2>/dev/null; EXTRA="$@"
printf 'request\trole\tboss\tlength\tcarry\tehps\toverheal\teh_per_mana\tmana_left\toom_at\ttank_died\tother_died\tdeath_floor\n' > "$OUT"
for r in druid_restoration druid_restoration_raid paladin_holy paladin_holy_raid priest_discipline priest_holy priest_holy_raid shaman_restoration shaman_restoration_raid; do
  case $r in *_raid) role=raid;; *) role=tank;; esac
  for b in onyxia broodlord chromaggus golemagg magmadar ragnaros vaelastrasz; do
    o=$($H -role $role -encounter tools/encounters/$b.json -carry -iters $IT $EXTRA tools/healbench/req/$r.json 2>/dev/null)
    len=$(echo "$o" | head -1 | grep -o '[0-9.]* s,' | head -1 | sed 's/,//')
    echo "$o" | grep "^| $r " | awk -F'|' -v r=$r -v role=$role -v b=$b -v len="$len" '{
      for(i=2;i<=NF;i++){gsub(/^ +| +$/,"",$i)}
      printf "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", r,role,b,len,$16,$4,$5,$6,$7,$8,$11,$12,$17}' >> "$OUT"
  done
done
