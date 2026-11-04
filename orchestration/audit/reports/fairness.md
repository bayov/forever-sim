# Fairness pass 30 (2026-11-04, master 0facd94)

Result: quiet (no new finding). 336 watched files changed since 6b480b0 (335 commits): `sim/core`, the item
database, standards 29 and 40, most classes' presets and the report requests, so this was a full pass: `audit1b.py`,
a full 5k race sweep on the loop's 0facd94 CLI and tree, and a second best-race sweep with the req/races overrides.
New baseline: 0facd94.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Phase 5 hit-trade regears (standards 4) | 74cc2b8e3, 3f4d44d4a, 3e4796fe7, c8a7b4ac3, 31c10b61c | Feral +25.8, Ret +11.8, Enhancement +8.6, Arcane +7.9, rogues +4.2 to +6.0, Destruction Fire +0.8 |
| Shaman hotfix push 113380 (Lightning Bolt, Chain Lightning, shocks) | 87922bb77 | Elemental +7.0, Mana Tide +6.9, Enh variants +1.4 to +6.5 |
| Unbridled Wrath on auto attacks only | 9173d087a | Arms +0.5 (request unchanged), Fury within 0.1 |
| Standards 29: hit-trade race gear for every race | 18f431e15 | report race still best (see below) |
| Tank / healer engine (mana threat, raid damage share, health costs tank-only, Berserking reads) | 49f02184c, b2c16f671, 318fc7efb, 87f28803b | no DPS build moves |
| Standards 40 (tank rulings) | f09015d81 | tanks only |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the 6b480b0 baseline: every hunter, the other mages and warlocks, priests, Balance, Fury DW / 2H
  identical to 0.1. Every move above traces to its own class's commit. Nothing shared leaked.
- Published against current: 13 builds off by more than 1.5 (Feral +26.5, SF Swords -11.9, SF Daggers -11.6, Ret
  +11.0, Enhancement +9.1, Arcane +7.9, Combat Daggers +7.9, Elemental +7.3, Subtlety +4.0, Mutilate -2.7, ...;
  variants Enh Mana Tide +7.4, Water Shield +7.3, Elemental Mana Tide +7.3). Already open as results-20261104-1 / -2
  (metas lag the regears; the phase 5 re-sim brief c40ede3b5 covers it). Independent confirmation here, not filed
  again. Note for that re-sim: the stale numbers misorder the published ranking (Feral shown behind both Furies,
  now ahead of them; Combat Daggers shown under Balance, now 2.4 above).
- Best race: report race is best everywhere. With the req/races overrides applied (`OVERRIDES=1 sweep.py`, the 27
  builds that have files) the overridden races gain 0.5 to 11 (High Order Fire +11.2, Balance Skyborne +10.6), but none
  passes the report race. Seal Fate Swords' Undead ties Night Elf and Gnome (0.2 under, noise).

## Ranking

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Enhancement 1168.5, Feral 1165.5, Fury 2H 1150.8, Fury DW 1135.7, Ret 1103.1,
Arms 1088.5, ... Smite 781.0 last. Top over bottom 1.68x. Full table: `orchestration/fairness_report.md` §1.

## Tooling

- `tools/audit/fairness/sweep.py`: `THREADS` (default 4) for loaded machines, and `OVERRIDES=1` to sim each race the
  way the report does (req/races rotation and hit-trade gear via the tree's `race_override.py`). The baseline log is
  without overrides; keep movement comparisons on plain sweeps.

## Open items (carried)

- None new. Next full sweep when `sim/` or the database changes again, and after the phase 5 report re-sim (expect
  the published gaps to close).
