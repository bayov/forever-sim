#!/bin/bash
# Writes the healing priest's shipped rotations and healbench requests from rots.py (run from the repo root).
set -e
P=orchestration/tools/healers/priest_tuning
HOLY_TANK="fhfirst=1 pws=0 fh=7,4 emergency=0.5 gate=0.7"
HOLY_RAID="pws=0.6 order=fh,poh,renew,gh renewshare=0.6 fh=7,5,3"
DISC_TANK="fhfirst=1 fh=7,4 emergency=0.5 gate=0.7"
(cd $P && python3 rots.py holy_tank ../../../../ui/healing_priest/apls/holy_tank_healer.apl.json $HOLY_TANK)
(cd $P && python3 rots.py holy_raid ../../../../ui/healing_priest/apls/holy_raid_healer.apl.json $HOLY_RAID)
(cd $P && python3 rots.py disc_tank ../../../../ui/healing_priest/apls/discipline_tank_healer.apl.json $DISC_TANK)
python3 $P/mk.py tools/healbench/req/priest_holy.json holy_tank holy $HOLY_TANK
python3 $P/mk.py tools/healbench/req/priest_holy_raid.json holy_raid holy $HOLY_RAID
python3 $P/mk.py tools/healbench/req/priest_discipline.json disc_tank disc $DISC_TANK
