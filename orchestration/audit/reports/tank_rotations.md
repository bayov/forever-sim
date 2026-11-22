# tank_rotations pass 24

Master 7a6d1beb9d (2026-11-22), against pass 23's baseline (master 8b5137c6d6). 1,000 iterations, default seed,
every tank request (13) on all seven boss files. One new finding (low).

## What changed

Gate: 8 watched files (8b5137c6d6 .. 7a6d1beb9d):

- `sim/core/debuffs.go` (bd7029410): under Forever Judgement of Wisdom pays only on landed swings. Every tank
  request but the three warrior presets carries the JoW debuff, so this changes their RNG draws (every bear,
  paladin and Warden row reseeds) and cuts paladin and Warden mana.
- `sim/core/attack.go`: a test-only helper.
- `sim/warrior/deep_wounds.go`, `execute.go`, `talents.go`: Deep Wounds sizing and no bleed off Thunder Clap
  crits, Execute's avoided refund and second Windfury roll, no Bloodthrill/Flurry off the totem's swing. The
  warrior presets take no Deep Wounds, Flurry or Bloodthrill and never Execute (Defensive Stance), and carry no JoW:
  all 32 warrior rows match pass 23 exactly.
- `sim/shaman/flame_shock.go` (haste shortens its GCD), `talents.go` (no Maelstrom stack off the totem): the
  Warden apls cast no Flame Shock; Maelstrom alone isn't visible behind the JoW reseed.
- `sim/paladin/item_sets_pve.go`: Crusader's Wrath shared with Soulforge; no tank wears either set.

No tank APL, request, rotopt tank template, encounter file, tankbench or tankscore change.

## Baseline

`ref/tank_bench/summary.tsv` rewritten for 7a6d1beb9d (`bench_summary.py` now takes master, date and pass as
arguments). Scoring means (score at the preset's weight; pass 23 -> 24):

| request | 23 | 24 | SafeDPS | healing needed |
|---|---|---|---|---|
| bear Max threat | 3121.0 | 3131.8 | 1850.8 -> 1853.3 | 642 -> 638 |
| bear Balanced | 2144.7 | 2130.0 | 1720.0 -> 1716.8 | 472 -> 473 |
| bear Max survival | 351.2 | 320.7 | 1324.3 -> 1319.3 | 288 -> 294 |
| paladin Max threat | 3248.8 | 3243.7 | 1739.0 -> 1734.5 | 272 -> 268 |
| paladin Balanced | 2809.3 | 2809.0 | 1710.8 -> 1706.8 | 244 -> 240 |
| paladin Max survival | 1780.2 | 1787.2 | 1629.3 -> 1627.2 | 202 -> 201 |
| Warden Max threat | 2621.0 | 2624.8 | 1707.8 -> 1708.7 | 860 -> 861 |
| Warden Balanced | 1288.0 | 1284.5 | 1555.5 -> 1551.0 | 662 -> 659 |
| Warden Max survival | -2192.8 | -2245.5 | 1497.2 -> 1489.7 | 628 -> 635 |
| warrior (all four) | unchanged | unchanged | | |

The bear and Warden moves are a reseed (the JoW roll no longer happens on avoided swings, so every later roll
shifts): single rows move up to about 1.6 times the unpaired difference's error and the means within it. The one
consistent move is the paladins on Ragnaros (not scored): Max threat 1094 -> 948 (± 37), Balanced -2260 -> -2747
(± 139), Max survival -12739 -> -13172 (± 214), healing needed +134 / +146 / +44, the expected effect of less JoW
mana over the long fight, and right per the change's log evidence. Not a finding; its rotation side is below.

## Rotation lines tried (paired abtest, scoring mean at the preset's weight)

| change | iters | seed 1 | seed 2 |
|---|---|---|---|
| paladin Balanced Consecration floor 0.6 -> any | 10,000 | +4.9 ± 1.0 | +3.9 ± 0.9 |
| ... floor 0.2 / 0.3 / 0.4 | 3,000 | +4.6 / +4.6 / +4.5 ± 1.6 | +3.4 / +3.4 / +3.5 ± 1.9 |
| ... floor 0.7 / 0.8 | 1,000 | -8.7 ± 4.3 / -15.0 ± 6.6 | |
| paladin Max survival Consecration floor 0.6 -> any | 3,000 | -3.2 ± 5.4 | +10.5 ± 4.8 |
| ... floor 0.4 / 0.8 | 3,000 | -2.4 / -21.5 | +11.3 / -18.8 |
| paladin Max threat Consecration gated at 30% mana | 1,000 | -0.0 ± 0.0 | |
| Warden Max threat 408345 mana floor 0.4 -> 0.2 / 0.6 | 1,000 | +1.3 ± 1.3 / +1.2 ± 2.8 | |
| Warden Balanced same line 0.4 -> 0.2 / 0.6 | 1,000 | -2.9 ± 6.7 / -6.0 ± 9.2 | |
| Warden Max survival same line 0.4 -> 0.2 / 0.6 | 1,000 | +6.1 ± 6.4 / +5.5 ± 14.4 | |

Only paladin Balanced's Consecration floor wins past the bar on both seeds: finding tank_rotations-20261122-1
(low, about 0.15%). docs/tank_paladin.md measured the same change at +0.9 / +0.7 on 2026-11-15, so the JoW change
made it. Execute for the warriors wasn't tried: the presets stay in Defensive Stance.

## Notes for the orchestrator

- The pass doc's open item (main-hand timing for Windfury) was done in pass 15; please remove it from
  passes/tank_rotations.md (asked in pass 23 too).
- The first tankbench build failed with "no required module provides package .../sim/core/proto"; running
  `protoc -I=./proto --go_out=./sim/core ./proto/*.proto` first fixed it (setup.sh installs protoc but doesn't run it).
