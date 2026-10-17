#!/bin/bash
# Usage: runclass.sh <jobsfile> <class>
# Runs sim.sh at 180 s and 5000 iterations for every line of <jobsfile> whose first field is <class>, printing "<tag> DPS ...".
# Job line format: <class> <spec> <gear> <apl> <talents> <race> <tag>, for example
#   fury fury level60_phase1_synthetic_fury level60_fury 30305013002-25053005112110501-2 Undead u2_fury
# Env: as sim.sh.
D=$(cd "$(dirname "$0")" && pwd)
grep "^$2 " $1 | while read c s g a t r tag; do echo "$tag $(bash $D/sim.sh $s $g $a "$t" $r 180 5000 $tag 2>&1 | grep -m1 -E 'DPS|SIMFAIL')"; done
