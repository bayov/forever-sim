# Rotations pass 73 (2026-11-23, master 5518f4873)

Quiet: the only rotation change is the mage re-tune, already measured by its worker; no new findings.

## What changed (7a6d1beb9..5518f4873)

- APL logic (`apl_logic_diff.py`): two files.
  - `ui/mage/apls/level60_frost.apl.json`: Ice Barrier (13033) refreshed below 2 s of Frigid Power (1326891), was 6 s.
  - `ui/mage/apls/level60_frostfire.apl.json`: Fire Blast (10199) at mana >= 40%, was 80%.
  - Both with talent changes (b8b89d6a8, "re-tune Frost and Frostfire for Ice Lance's 0.11 coefficient").
- Sim `.go` (non-test): `sim/common/forever_armor_uses.go` and `item_effects.go` (Helm of Fire, Circle of Flame, Robe of
  the Archmage, Headmaster's Charge, Axe of the Deep Woods, Flame Wrath), `sim/core/attack.go` and
  `apl_actions_timing.go` (an empty weapon slot doesn't swing; a wait on NeverExpires doesn't overflow). None of them
  moves a raid preset (baseline below). tools/rotopt Phase 0 files are outside the presets.

## Baseline (3000 it, seed 424242), new file `ref/rotations/baseline_5518f48_3000it_seed424242.tsv`

| preset | old | new | move |
|---|---|---|---|
| mage frost | 835.1 | 839.6 | +0.53% |
| mage frostfire | 1006.9 | 1009.7 | +0.28% |

The other 66 presets are unchanged to 0.05 DPS.

## Spot-check of the merged re-tune

Frost with Ice Barrier back at 6 s (B) vs master's 2 s (A), paired `ab.py`, 20k it, 8 blocks, B-A at 120 / 180 / 300 s:

| S0 1e6 | S0 2e8 | pooled |
|---|---|---|
| -0.75 / -0.98 / -0.68 | -0.45 / -0.54 / -0.61 | -0.60 / -0.76 / -0.65 |

A at 180 s: 839.5. The 2 s gate holds (the worker's +3.9 includes the talent change). Frostfire's Fire Blast gate was
not re-run; the worker confirmed it on two seeds at 40k.

## Open findings

- rotations-20261122-1 (Shock Ret HoW/Holy Strike order) and rotations-20261112-1 (rogue poison swap): not landed
  (no paladin or rogue APL/req change in the range). Still open.

## Optional

None new. The pass-doc list of optional results stands.
