# Results pass 58, master 224b8a4e0 (2026-10-28)

Trigger: 9 new race overrides in tools/dps_report/req/races/ (1bffad948, c55e6f38e, 19257c429, 7d07cf709, 4c9f2d32c, 84be79f37) and the builds.json republish. No change under sim/, proto/, mksettings or assets/database/, so drift and unread fields are unchanged from pass 57.

| Check | Result |
|---|---|
| Shared run, 58 requests, 20k seed 1 vs builds.json | all 58 within 0.15 |
| Metas' published.dps vs run | 34 of 58 still off (open finding results-20261027-1) |

Race overrides, paired 20k seed 1 (same request with and without the override file, `$AUDIT_WORK/results/pair.py`):

| Build, race | plain | override | gain | claimed |
|---|---|---|---|---|
| mage__frostfire High Order | 1057.28 | 1076.05 | +18.76 | 18.8 |
| warrior__fury Night Elf | 968.09 | 977.41 | +9.32 | 9.32 |
| mage__arcane High Order | 891.51 | 896.43 | +4.93 | 4.93 |
| shaman__elemental-level60 Troll | 762.57 | 767.43 | +4.86 | 4.86 |
| shaman__elemental-level60-mana-tide Troll | 747.38 | 752.06 | +4.68 | 4.68 |
| warlock__demonology Gnome | 1045.85 | 1049.95 | +4.10 | 4.1 |
| warlock__destruction Gnome | 994.63 | 998.07 | +3.44 | 3.44 |
| warlock__affliction Gnome | 1012.66 | 1015.92 | +3.26 | 3.26 |
| warlock__destruction_fire Gnome | 977.94 | 979.14 | +1.21 | 1.21 |

All 9 gains reproduce exactly; destruction_fire's +1.21 is 8.6 SE (se 0.14 in its file).
Race files hold no full request (keys: rotation, note, gain, how), so runreq.sh cannot sim them directly; use race_override.apply as in pair.py.
No new finding.
