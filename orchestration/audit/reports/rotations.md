# Rotations pass 46 (quiet)

- Master: fa865a3b8b7a (from 5a6e9c6a6), audited 2026-11-02T06:53:11Z. Client 1.60.1.70707, unchanged.
- Baseline unchanged: `ref/rotations/baseline_5a6e9c6_3000it_seed424242.tsv` (no sim change the presets reach).

## What changed

| Change | Effect on Level 60 rotations |
|---|---|
| Deep Holy Shock Ret build dropped (231d5492a, user ruling 2026-11-02): `level60_shock{,_undead,_crusader_duty,_wisdom_duty}.apl.json` and their reqs (plus the Dwarf race req) deleted | Deletions only, so no live APL logic changed |
| `sim/compsearch/pool.go`, `drops.json` | Comp search tables only; raid presets don't reach them |
| `sim/raid_forever_test.go`, `report_hashes.txt`, `tools/rotopt/*` | Tests and tooling only |
| `*.meta.json` refresh | Metadata only |

## Checks

- `apl_logic_diff.py`: 4 files touched, all whole-file removals.
- Nothing dangling: the remaining `shock` references all point at the kept `paladin__level_60_shock_ret*` build, whose
  3 APLs and 3 reqs are present. The shared run sims all 3 (965 to 975 DPS).
- Open findings: none.

## Optional (under the bar)

Carried over from the pass doc, not re-measured: Rogue Mutilate window 0.5 s, Fury-BS HS 55, hunter end-of-fight gates,
Shadow Mind Blast time-left gate, Destruction Fire DoT gates, Water Shield Magma gate.
