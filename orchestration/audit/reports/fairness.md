# Fairness pass 31 (2026-11-05, master a079415)

Result: quiet (no new finding). 42 watched files changed since 0facd94: `sim/core` (heal threat in a score's replay,
hit-step items left out of the AllItems goldens), explosives in forms, Feral's Ferocious Bite hotfix, the Survival and
Ret rulings, the tank presets. Full pass: `audit1b.py`, a full 5k race sweep on the loop's a079415 CLI and tree, and an
explosives probe on the specs without the sapper. New baseline: a079415.

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Sapper and Land Mine usable in a form (client NOT_SHAPESHIFTED flag) | 22a7d7913 | Feral requests take the sapper; Feral +4.1 with the FB hotfix |
| Ferocious Bite extra Energy damage (hotfix 113403) | ff56f8ddc | Feral (above) |
| Survival takes Rapid Killing for Resourcefulness (user ruling) | b682e47a8 | Survival -0.8, wolf -0.9, tallstrider +1.1 (noise) |
| Ret twist on hit-3 is the paladin default (user ruling) | b682e47a8 | UI default only; report requests unchanged |
| Heal threat reads no health in a replay; golden test generator skips hit-step items | 464deda7b, aa7cd24dc | no DPS build moves (goldens moved, report requests did not) |
| Tank rounds 4-5, standards 40 | several | tanks only |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Movement against the 0facd94 baseline: only Feral and the three Survival builds moved; everything else identical to
  0.1. Both trace to their own class's commits. Nothing shared leaked.
- Explosives in forms: with the form block gone, I checked the specs whose requests carry no sapper (5k, seed 1):
  Balance +0.0 with the sapper, -4.9 with the Dark Bomb; Fire +0.0 / -6.7; MM +0.3 with the sapper (noise). "Skip when
  it loses" (standards 11) still holds; no unevenness.
- Published against current: the same rows as pass 30 off by more than 1.5, Feral now +30.6 (Ret +11.0, SF Swords
  -11.9, SF Daggers -11.6, Enhancement +9.1, Arcane +7.9, Combat Daggers +7.9, Elemental +7.3, Subtlety +4.0, Mutilate
  -2.7). Already open as results-20261104-2 (and -20261105-1 for the mksettings sapper drift); independent confirmation
  here, not filed again. The stale metas show Feral behind Enhancement and both Furies; it is 4th now.
- Best race: report race is best everywhere (Seal Fate Swords' Undead 0.2 under Night Elf, noise).

## Ranking

Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Feral 1169.6, Enhancement 1168.5, Fury 2H 1150.8, Fury DW 1135.7, Ret 1103.1,
Arms 1088.5, ... Smite 781.0 last. Top over bottom 1.68x. Full table: `orchestration/fairness_report.md` §1.

## Open items (carried)

- None new. Next full sweep when `sim/` or the database changes again, and after the metas are refreshed (expect the
  published gaps to close).
