#!/bin/bash
# run2.sh: the Zanza / Blasted Lands / rum slots (standards item 11 ruling, 2026-11-04) for the Warden and the bear, 10k on seeds 1 and 2.
R=orchestration/tank_handoff/shaman/round5
REQ=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/r5req
Z="zanza_spirit zanza_sheen zanza_swift zanza_roids zanza_scorpok zanza_lung zanza_gizzard zanza_cortex rum stout"
for p in max_threat balanced max_survival; do
  python3 $R/gen.py shaman_warden $p $REQ/shaman_warden $Z
  python3 $R/gen.py druid_bear $p $REQ/druid_bear $Z
done
for sd in 1 2; do
  for p in max_threat balanced max_survival; do
    bash $R/ab.sh shaman_warden $p 10000 $sd $Z
    bash $R/ab.sh druid_bear $p 10000 $sd $Z
  done
done
echo DONE
