# tank_rotations pass 34

master f9fa79917b, 2026-11-26. Previous pass on ab3fff7007; baseline compared against `ref/tank_bench/summary.tsv`
(phase0-tankheal 95962a7d3, merged in a589203b7). Result: quiet (no new findings; four open findings resolved).

## What changed

The phase0-tankheal merge: every tank request on Phase 0 gear and consumables, bear Balanced and Max threat without
Demoralizing Roar (09709d889, Balanced re-searched b7b411a01), bear Max survival Lacerate under 10 s (a2f62f579),
requests cast their worn on-use trinkets (0741191ab, ecf1cbb50, 2f20719e4, 6adc8f19c), Warden races Troll/Tauren
(3ea8c0ecc), paladin consumables (3ae9da70a, 52572a1d7), Water Shield's 4th globe with Totem of Sustaining (7de8438a6).
Warrior sim code changes in the range are leveling-only (passes 32 and 33).

## Baseline (1,000 iterations, default seed, scoring mean at the preset's weight)

| request | score (scoring mean) | was | SafeDPS | was | healing needed | was |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 3771.5 | 3645.8 | 2955.2 | 2776.8 | 758.5 | 687.2 |
| druid_bear_max_survival | 1048.0 | 1015.3 | 1914.3 | 1875.3 | 344.8 | 339.5 |
| druid_bear_max_threat | 5606.7 | 5386.8 | 3229.7 | 3115.5 | 918.0 | 909.2 |
| paladin_protection_balanced | 2965.3 | 2914.8 | 2097.7 | 2080.2 | 441.0 | 448.2 |
| paladin_protection_max_survival | 906.5 | 800.2 | 1676.8 | 1825.8 | 292.5 | 344.0 |
| paladin_protection_max_threat | 3936.3 | 3912.5 | 2205.3 | 2218.3 | 514.7 | 572.3 |
| shaman_warden_balanced | 2022.7 | 2026.2 | 1977.8 | 1980.0 | 698.3 | 698.5 |
| shaman_warden_max_survival | -1212.8 | -1337.0 | 1623.5 | 1675.3 | 548.7 | 572.8 |
| shaman_warden_max_threat | 3428.0 | 3418.0 | 2124.7 | 2114.7 | 893.2 | 884.2 |
| warrior_protection | -399.2 | -399.2 | 854.5 | 854.5 | 755.5 | 755.5 |
| warrior_protection_balanced | 3360.5 | 3063.7 | 2418.3 | 2300.5 | 500.0 | 518.3 |
| warrior_protection_max_survival | 1218.8 | 1095.5 | 1763.0 | 1770.7 | 265.7 | 281.2 |
| warrior_protection_max_threat | 4725.8 | 4467.5 | 2735.8 | 2622.8 | 743.5 | 776.7 |

Attribution: the old requests (from a589203b7) re-run on f9fa79917b reproduce the old summary's SafeDPS exactly on 84 of 91
rows; the 7 that differ are bear Max survival, whose request moved between 95962a7d3 and a589203b7 (Lacerate 10 s).
So every move is a request change (gear, Roar, trinkets, races, consumables), none an engine change. New
`ref/tank_bench/summary.tsv` written.

## Changes tried (paired abtest, 3,000 iterations, B minus preset)

| change | preset | seed 1 | seed 2 | per boss |
|---|---|---|---|---|
| delete Thunder Clap (priority 14) | warrior Balanced, w = 1 | -3.3 ± 12.1 | -15.2 ± 10.6 | Chromaggus -258 / -273, Vaelastrasz +192 / +183, Ragnaros -15 / -74 |
| Consecration at any mana (priority 7) | paladin Balanced, w = 1 | 0.0 ± 0.0 | 0.0 ± 0.0 | bit-identical: the 60% floor never binds on the Phase 0 set |

Both old findings (20261125-2, 20261122-1) are resolved: neither change wins on master.

## Findings

- Resolved: 20261125-1 (09709d889), 20261125-2 (no longer gains), 20261125-3 (a6085ef3e), 20261122-1 (moot).
- Still open, user calls: 20261104-11 (warrior Max survival trinket hold), 20261104-3 (paladin Bulwark 13 s). Not re-tested.

## Open item not done

The Windfury main-hand timing item (pass doc, user request 2026-11-15) is still not done: it needs new APL
conditions for swing timing across four tank specs on two seed sets, more than one pass budget. Left in the pass doc.
