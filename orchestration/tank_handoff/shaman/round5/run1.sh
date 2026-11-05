#!/bin/bash
# run1.sh: the Warden consumable search (task 1) and the natural-flask A/B (task 2), 10k on seeds 1 and 2.
R=orchestration/tank_handoff/shaman/round5
REQ=/tmp/claude-0/-home-user-forever-sim/9040db86-6de0-5226-9d88-b0db7baa031d/scratchpad/r5req
W1="flask_titans flask_wisdom flask_supreme flask_chromatic agi_greater agi_none armor_superior armor_none hp_forever hp_none str_giants str_none ap_firewater ap_none sp_arcane fire_greater nature_power mageblood int_greater trolls_major pot_stoneshield pot_healing pot_none conj_rune chili_off mh_elemental mh_dense mh_manaoil mh_wizardoil"
NAT="flask_prec flask_acc flask_aggr flask_swift"
for p in max_threat balanced max_survival; do
  python3 $R/gen.py shaman_warden $p $REQ/shaman_warden
  python3 $R/gen.py druid_bear $p $REQ/druid_bear $NAT stone_elemental
done
for sd in 1 2; do
  for p in max_threat balanced max_survival; do
    bash $R/ab.sh shaman_warden $p 10000 $sd $W1 $NAT
    bash $R/ab.sh druid_bear $p 10000 $sd $NAT
  done
done
echo DONE
