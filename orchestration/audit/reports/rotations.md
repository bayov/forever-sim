# Rotations pass 44 (2026-11-01)

- **Master:** c23ef0440 (last audited dd053d24c). Client 1.60.1.70707.
- **Result:** findings. One new, low: rotations-20261101-1 (Frost mage, 4th Arcane Blast stack above 70% mana).

## What changed

- `apl_logic_diff.py dd053d24c c23ef0440`: 0 APL files touched. Only `.meta.json` req files changed.
- Sim changes the presets reach:
  - `sim/core/rage.go` (5b2eec6): a two-hand white crit doubles the whole swing's Rage.
  - `sim/hunter/pet.go`, `pet_forever.go`: no Pet Aggression, pet swing = creature dps less the pet's AP.
  - `sim/mage/*` (b79e033, 98e819d): Arcane Blast 12% Int / +100% mana a stack / 18 s; Ice Lance x4 on
    Frozen; Frigid Power 15%, shielded 5%; Winter's Chill 20% on Cone of Cold.
  - `sim/warlock/talents.go`: Improved Drains off Wrack. `sim/core/buffs.go`: Moonkin Aura / Leader of
    the Pack 3% crit, not stacking. `sim/common/item_effects.go`: Kang's Rend proc. Righteous Fury threat.

## Regression

`baseline.py` on the shared CLI (3,000 it, seed 424242) against `baseline_dd053d2`. No req .json changed,
so every move is the sim. Rows over 1 DPS:

| preset | old | new | diff |
|---|---|---|---|
| hunter__beast-mastery | 1124.7 | 1044.5 | -80.1 |
| hunter__beast-mastery-wolf / -tallstrider | 1068.8 / 1054.9 | 1001.6 / 987.7 | -67.3 / -67.2 |
| hunter__survival / marksmanship | 1069.3 / 1108.3 | 1023.8 / 1063.0 | -45.5 / -45.3 |
| other pet hunters (4) | | | -38.1 to -38.2 |
| mage__arcane | 911.5 | 894.4 | -17.1 |
| warrior 2H presets (7) | | | +10.2 to +14.7 |
| mage__frost | 909.4 | 930.0 | +20.6 |
| mage__frostfire | 1048.9 | 1079.3 | +30.3 |

Warlocks, druids, paladins: under 1 DPS. New baseline `baseline_c23ef04_3000it_seed424242.tsv` (replaces dd053d2).

## A/B checks (paired, 20k it in 8 blocks; B-A at 120 / 180 / 300 s)

| preset | variant | S0 = 1e6 | S0 = 2e8 | verdict |
|---|---|---|---|---|
| mage__frost | AB cap 3 -> 4 | +4.6 / +2.7 / -15.7 | | trade-off (mana) |
| mage__frost | 4th stack if mana > 0.7 | +2.2 / +1.5 / +0.3 | +3.2 / +2.1 / +0.7 | **finding** |
| mage__frost | 4th stack if mana > 0.6 | +2.6 / +2.0 / -0.2 | +3.5 / +2.4 / +0.6 | alternative |
| mage__frost | stack-refresh gate 9 s -> 12 s | 0 / 0 / 0 | | inert |
| mage__arcane | AB cap 4 -> 3 | -31.5 / -31.1 / -30.3 | | keep 4 |
| mage__frostfire | Ice Lance on any Fingers of Frost (not ==1) | -4.5 / -23.0 / -20.6 | | keep |
| warrior__arms | Heroic Strike 90 -> 70 / 60 | -9.5 / -10.3 / -8.5; -19.8 / -21.5 / -18.5 | | keep 90 |
| warrior__arms | Heroic Strike 90 -> 95 | +0.9 / +0.4 / +0.3 | +0.2 / +0.6 / +0.1 | optional |
| warrior__fury_2h | Heroic Strike 90 -> 70 | -7.8 / -10.3 / -6.8 | | keep |
| warrior__fury_2h | Heroic Strike 90 -> 95 | +0.0 / +0.4 / -0.9 | | no |
| warrior__fury_2h | Slam 45 -> 35 | -1.2 / -1.0 / -0.9 | | keep |

The crit Rage fix leaves the 2H Heroic Strike thresholds at their optimum.

## Checked: Ice Barrier for Frost now that Frigid Power is 15%

None of the mage presets talent Ice Barrier (a test cast never fires). A quick 31-Frost build with Ice
Barrier (`0502052031001-1-253500030100301251`, Frost Warding 2, Cold Snap, Winter's Chill 5, Ice Barrier;
Arcane Mind and Instability dropped) recast on Frigid Power gave 899.1 against 930.0 (3,000 it, seed
424242). Not a finding. A full talent search wasn't run (out of this pass's scope).

## Optional (under the bar)

- Arms Heroic Strike at 95 rage: +0.4 / +0.6 at 180 s on the two seed sets.
- The Arcane APL's Arcane Blast note also still says "15% of Intellect ... for 15 sec" (now 12%, 18 s); wording only.

## Open items carried

- Hunter pet rework (-38 to -80): no APL lever found. The pet's own abilities aren't in the APLs.
- Older open items in `passes/rotations.md` (Balance Night Elf, Elemental Mastery hold, Warlock duty
  variants) stay open.
