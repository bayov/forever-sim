#!/bin/bash
# search.sh <talents|knobs> <label> <request> <tank> <talents> [race] : rotopt as round 3 ran it (six scoring bosses, 1,000 a
# candidate, 3,000 to confirm, Bulwark frozen at 10 s); output <kind>_<label>.txt, knob APLs to aplout/
cd /home/user/forever-sim; R=orchestration/tank_handoff/paladin/round5
ENC=tools/encounters/magmadar.json,tools/encounters/golemagg.json,tools/encounters/onyxia.json,tools/encounters/broodlord.json,tools/encounters/vaelastrasz.json,tools/encounters/chromaggus.json
kind=$1; label=$2; req=$3; tank=$4; tal=$5; race=${6:-Undead}
ro="/tmp/claude-0/paladin_r5/rotopt -template paladin_tank -settings $req -tank $tank -tank-encounter $ENC -iters 1000 -confirm-iters 3000 -freeze bulwarkAt -race $race -talents $tal"
if [ $kind = talents ]; then $ro -talent-search > $R/talents_$label.txt 2>&1
else mkdir -p $R/aplout/$label; $ro -out $R/aplout/$label -name level60_$label > $R/knobs_$label.txt 2>&1; fi
