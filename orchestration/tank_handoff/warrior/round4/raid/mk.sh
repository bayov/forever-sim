#!/bin/bash
# mk.sh <variants file> <base dir> <out dir> : writes <out>/<name>_<preset>.json for each line "name arg..." (mkvar.py) and preset
R=$(dirname $0); mkdir -p $3
while read -r n args; do [ -z "$n" ] || [ "${n:0:1}" = "#" ] && continue
 for p in max_threat balanced max_survival; do python3 $R/mkvar.py $2/warrior_protection_$p.json $3/${n}_$p.json $args; done
done < $1
