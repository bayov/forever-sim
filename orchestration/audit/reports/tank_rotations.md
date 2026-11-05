# tank_rotations: pass 5 (2026-11-05, master bba052f75)

Due on 1 watched file changed since a07941531: `tools/tankbench/main.go` (cb4065412, the tank and heal race
sims). It adds three opt-in flags and nothing else: `-first-seed` (raw seed of fight 0), `-labeled`
(`SimOptions.UseLabeledRands`) and `-vectors` (appends per-fight SafeDPS, survive HPS, DTPS, TPS as JSON lines,
setting `SaveAllValues`). Unset, each leaves the request exactly as before. The same range adds
`RaidMetrics.FightDeaths` (sim/core/raid.go `doneHealingFight`, combined in sim_concurrent.go): a per-fight
bitmask (1 tank died, 2 other died), metrics only, no random draws. No tank APL, request, rotopt template,
encounter file, tankscore or tank ability changed.

## Baseline (ref/tank_bench/summary.tsv, header now bba052f75)
Every request on every boss (13 x 7 = 91 rows), 1,000 iterations, default seed, no new flags. All 91 rows equal
pass 4's ref to the decimal (`diff` of the data lines is empty), so the tankbench change moves nothing at its
defaults. Scoring means at own weight are therefore pass 4's:

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
| warrior_protection_max_survival | 3 | 2139 | 2059 | 336 | 240 | 0.0 |
| warrior_protection_max_threat | 0.333 | 4652 | 2535 | 415 | 427 | 0.0 |

## Rotation lines
No rotation, request or template changed since pass 4, which probed 36 one-line A/Bs (3,000 iterations, paired)
on these exact APLs and found nothing new. Not re-probed this pass.

## Open findings
- 3 (paladin Bulwark 13 s) and 11 (warrior Max survival armor trinket held 35 s): user calls, unchanged on
  master; not re-tested.

## Conclusion
Quiet. No new findings.

Note for the next pass: the shared `bench_summary.py` writes straight to the ref with a hard-coded header; this
pass ran it with the output redirected to `$AUDIT_WORK/tank_rotations/new.tsv`, diffed, then edited the ref's
header line only.
