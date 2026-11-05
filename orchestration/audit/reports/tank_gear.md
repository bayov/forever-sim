# tank_gear pass 4 (master a079415316, since dc303b0)

Watched changes: standards.md 40 (tanks get the full raid consumable set; no gear rule), paladin gear set: boots enchant 929 -> 1887
(Greater Agility, live boots enchant in db.json), Max threat's separate set deleted (round 5 takes Greater Agility boots at every preset).

`python3 tools/audit/tank_gear/paladin.py`: every piece within 1% linear (99.1-100.6) and 6% norm (96.1-104.2); set total 841.8 vs 844.1;
Stamina 32.6 (32.4 +-1.5), all other shares within 0.2 of hit-adjusted targets; hit 5.0 + 3 Precision = 8%; Defense +7 2pc set bonus counted.
Gear set equals its requests (990710 boots 1887). Warrior, bear and Warden sets unchanged, not re-run.

Note (not a finding): docs/tank_paladin.md:249 still names the deleted level60_phase1_synthetic_max_threat.gear.json.
