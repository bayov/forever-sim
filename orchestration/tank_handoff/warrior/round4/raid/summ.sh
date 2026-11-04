#!/bin/bash
# summ.sh <glob> : one line per ab file: name, scoring mean diff ± paired error (B - A)
cd $(dirname $0)/ab; for f in $1; do echo "$f $(grep scoring_mean $f | awk '{print $2, "±", $3}')"; done
