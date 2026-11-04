#!/bin/bash
R=/home/user/forever-sim/orchestration/tank_handoff/paladin/round5
$R/search.sh talents mt $R/var/base_max_threat.json max-threat 05005-25300102413001531-5023 &
$R/search.sh talents bal $R/var/base_balanced.json balanced 04-25300102513501531-5023 &
wait
$R/search.sh talents bal_from_mt $R/var/mtbuild_balanced.json balanced 05005-25300102413001531-5023 &
$R/search.sh talents ms $R/var/base_max_survival.json max-survival 04-25300102513501531-5023 &
wait
