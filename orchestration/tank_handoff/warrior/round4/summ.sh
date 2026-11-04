#!/bin/bash
# summ.sh <glob-suffix> : one line per file: scoring_mean diff ± se
for f in s*_*$1.txt; do l=$(grep scoring_mean $f); echo "$f $(echo $l | awk '{print $2, "±", $3}')"; done
