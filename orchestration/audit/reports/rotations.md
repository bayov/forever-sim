# Rotations pass 74 (2026-11-24, master 566067a39)

Quiet: no Level 60 APL changed; three sim changes move presets, and none of them changes the best rotation on the
variants tried. No new findings.

## What changed (c78d983ac..566067a39)

- APL logic (`apl_logic_diff.py`): no APL files touched. No `tools/dps_report/req` change.
- Sim `.go` (non-test):
  - `sim/druid/rip.go`, `rake.go`, `tuning_70753.go` (be773f5eb): Rip tick AP share 0.0125 -> 0.0114 a combo point,
    Rake tick 0.06875 -> 0.065, from wowsims/forever's logs (user ruling 2026-11-22).
  - `sim/warrior/talents.go`, `reckless_swings.go`, `heroic_strike_cleave.go` (cd71e78c2): Heroic Strike and Cleave spend
    a Flurry charge again (EV-0138), Reckless Swings charge on every HS/Cleave cast, landed or not; Unbridled Wrath no
    longer procs on Shield Slam, Revenge or Shield Bash.
  - `sim/hunter/aspects.go`, `pet.go` (b82ad9a1e): the pet's summon snapshot takes the prepull Aspect of the Hawk.
  - tools/rotopt Phase 0 files are outside the presets.

## Regression (old CLI c78d983ac vs new, same req files, 3000 it, seed 424242), moves over 0.5 DPS

| preset | old | new | move |
|---|---|---|---|
| druid feral-cat | 1186.4 | 1171.4 | -15.1 (-1.27%) |
| druid feral-cat-duty-ff | 1179.6 | 1164.7 | -14.9 |
| warrior fury | 1113.1 | 1109.4 | -3.6 |
| warrior fury-duty-battle-shout | 1110.8 | 1105.4 | -5.5 |
| warrior fury-duty-sunder | 1073.1 | 1071.6 | -1.5 |
| warrior fury_2h / -duty-sunder | 1129.4 / 1092.3 | 1128.9 / 1091.0 | -0.6 / -1.3 |
| hunter beast-mastery (+tallstrider, wolf) | 1111.9 | 1113.3 | +1.2 to +1.4 |
| hunter marksmanship (+tallstrider, wolf) | 1199.1 | 1199.9 | +0.7 to +0.8 |

New baseline `ref/rotations/baseline_566067a_3000it_seed424242.tsv` matches this table (old file removed).
The larger moves in the loop's `vs_published.txt` (mage, warlock, rogue, shaman) predate c78d983ac or come from gear,
not from this range's sim code.

## A/B on the moved rotations (paired `ab.py`, 20k it, 8 blocks, S0 1e6, B-A at 120 / 180 / 300 s)

| preset | variant B | B-A | verdict |
|---|---|---|---|
| feral-cat (A 1171.4 at 180) | no Rip (Bite only) | -127.8 / -131.8 / -136.4 | loses |
| feral-cat | no Rake | -91.0 / -92.8 / -92.4 | loses |
| feral-cat | Rip only with > 12 s left (was 8 s) | -2.61 / -1.67 / -1.07 | loses |
| feral-cat | Rip with > 6 s left | -0.10 / -0.08 / +0.02 | flat |
| feral-cat | Rake with > 9 s left (was 6 s) | -1.52 / -0.96 / -0.47 | loses |
| feral-cat | Rake with > 4 s left | +0.40 / +0.28 / +0.16 | under the bar (optional) |
| fury (A 1109.6 at 180) | Heroic Strike at rage >= 50 (was 60) | -1.63 / -3.09 / -2.31 | loses |
| fury | Heroic Strike at rage >= 70 | +0.35 / -0.46 / -1.42 | loses |

SE 0.01-0.6. No candidate reached the bar, so no second seed set was run. Hunter moves are a stat snapshot, no
rotation lever.

## Optional (under the bar)

- Feral: Rake end-of-fight gate 6 s -> 4 s, about +0.3 at 180 s (one seed set).

## Open findings

- rotations-20261122-1 (Shock Ret HoW/HS order) and rotations-20261112-1 (rogue poison swap): not landed; still open.

## Still open from the pass doc

- The warrior HS/Cleave re-search item now has one more reason (HS/Cleave spend Flurry again); the quick 50/70 check
  above keeps 60 on Fury. Fury 2H and the duty variants were not re-checked (moves of 0.6-5.5).
- Windfury main-hand timing, paladin hasted-GCD re-time and the other open items are unchanged.
