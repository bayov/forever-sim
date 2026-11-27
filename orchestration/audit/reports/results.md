# results: branch pass, phase0-cleanup 782ad6089 (2026-11-27)

Scoped by hand to the Talisman of Ascendance wearers (19 requests: 13 warlock, 4 Balance, 2 Elemental); state not advanced.
CLI built at HEAD, `run_reqs.py`, 20k, seed 1, 180 s.

| build | HEAD | builds.json | meta.published |
|---|---|---|---|
| warlock__affliction | 962.83 | 962.5 | 962.47 |
| warlock__affliction-duty-curse | 948.15 | 948.1 | 953.4 |
| warlock__affliction-duty-recklessness | 942.66 | 942.6 | 948.24 |
| warlock__affliction_shadow_bolt | 818.63 | 818.4 | 820.46 |
| warlock__demonology | 874.56 | 873.9 | 876.96 |
| warlock__demonology-duty-curse | 864.64 | 864.4 | 868.87 |
| warlock__demonology-duty-recklessness | 859.00 | 858.8 | 863.52 |
| warlock__destruction | 866.64 | 866.3 | 868.02 |
| warlock__destruction-duty-curse | 856.14 | 856.8 | 860.52 |
| warlock__destruction-duty-recklessness | 851.03 | 851.2 | 855.06 |
| warlock__destruction_fire | 824.24 | 823.9 | 825.47 |
| warlock__destruction_fire-duty-curse | 815.58 | 815.6 | 818.07 |
| warlock__destruction_fire-duty-recklessness | 810.83 | 810.6 | 812.9 |
| druid__balance (4 builds) | +0.10 to +0.39 vs meta | | |
| shaman__elemental (2) | -0.74 / -0.12 vs meta | | |

- The published dashboard matches HEAD within the report's 180 +/- 20 s spread (at most 0.7). The warlock drops
  (1.6 to 5.6 against the old publish, biggest on the duty rows that cast an extra curse) are explained by 413f06cb:
  Life Tap and curses now spend charges.
- New: results-20261127-1 (low). The 13 warlock metas keep the pre-fix published.dps.
- Not checked this pass: drift, unread fields, and the other 49 requests (no other sim change on the branch).
