# tank_rotations: pass 6 (2026-11-08, master 9ba00c2dc)

Due on 11 watched files changed since bba052f75 (all `sim/shaman/*`). The range (8 commits) also touches the
engine: whole-point spell damage rolls (2a60326a1, `sim/core/sim.go`, `cast.go`), Forever's downranking penalty on
heals only (0193721bc, `sim/core/downrank.go`), Windfury and Flametongue Totem sharing the weapon's totem slot
(9ba00c2dc, `sim/core/totem_weapon_buffs.go`, shaman totems and imbues), and the Stormstrike mark after the damage
calc (5a870eccc). No tank APL, tankbench request, rotopt tank template, encounter file, `tools/tankbench/main.go`
or `tools/tankscore` changed.

The totem slot change can't reach the tank benches: all 13 requests raid-buff `TotemWeaponBuffWindfury` only,
none Flametongue Totem.

## Baseline (ref/tank_bench/summary.tsv, header now 9ba00c2dc)
Every request on every boss (13 x 7 = 91 rows), 1,000 iterations, default seed. 96 of 104 summary rows (incl.
scoring means) equal pass 5 to the decimal; the other 8 move by 1 point (score, SafeDPS or TPS), far inside their
errors (score_err 31 to 59): the whole-point rolls reshuffle a few spell hits. All Warden shaman rows are unchanged.

| request | w | score | SafeDPS | DTPS | healing needed | death floor |
|---|---|---|---|---|---|---|
| druid_bear_balanced | 1 | 2498 | 1883 | 584 | 456 | 0.0 |
| druid_bear_max_survival | 3 | 238 | 1649 | 544 | 372 | 0.0 |
| druid_bear_max_threat | 0.333 | 3492 | 2024 | 610 | 614 | 0.0 |
| paladin_protection_balanced | 1 | 2835 | 1724 | 314 | 241 | 0.0 |
| paladin_protection_max_survival | 3 | 1910 | 1611 | 295 | 180 | 0.0 |
| paladin_protection_max_threat | 0.333 | 3249 | 1731 | 305 | 254 | 0.0 |
| shaman_warden_balanced | 1 | 1786 | 1857 | 677 | 694 | 0.0 |
| shaman_warden_max_survival | 3 | -1988 | 1788 | 686 | 674 | 0.0 |
| shaman_warden_max_threat | 0.333 | 3170 | 1949 | 721 | 795 | 0.0 |
| warrior_protection | 1* | -14 | 1048 | 637 | 757 | 0.0 |
| warrior_protection_balanced | 1 | 3880 | 2386 | 376 | 315 | 0.0 |
| warrior_protection_max_survival | 3 | 2138 | 2059 | 336 | 240 | 0.0 |
| warrior_protection_max_threat | 0.333 | 4652 | 2535 | 415 | 427 | 0.0 |

Rows that moved (pass 5 -> pass 6): druid_bear_max_survival onyxia score -315 -> -314; paladin_protection_max_threat
ragnaros score 993 -> 994; warrior_protection_balanced vaelastrasz SafeDPS 2533 -> 2532; warrior_protection_max_survival
golemagg TPS 2004 -> 2003, magmadar score 3248 -> 3247 and SafeDPS 2035 -> 2036 (plus the three scoring means).

## Rotation lines
No rotation, request or template changed since pass 4, which probed 36 one-line A/Bs (3,000 iterations, paired)
on these APLs; the engine moved no row by more than a point, so nothing there can have flipped. Not re-probed.

## Open findings
- 3 (paladin Bulwark 13 s) and 11 (warrior Max survival armor trinket held 35 s): user calls, unchanged on
  master; not re-tested.

## Notes for the next pass
- `tools/audit/tank_rotations/bench_summary.py` reads `/tmp/audit-work/tank_rotations/base/` (hardcoded) and writes
  the ref directly with a stale header line; fix the header by hand after running it. Export `AUDIT_WORK` in every
  shell (the env from `source tools/audit/setup.sh` doesn't carry across tool calls), and generate the Go protos
  (`protoc -I=./proto --go_out=./sim/core ./proto/*.proto`) before `go build`.

## Conclusion
Quiet. No new findings.
