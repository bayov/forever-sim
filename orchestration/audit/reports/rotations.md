# Rotations pass 48 (quiet: no findings)

- Master: a0b306984 (from 796495d6f), audited 2026-11-03. Client 1.60.1.70707, unchanged.
- New baseline: `ref/rotations/baseline_a0b3069_3000it_seed424242.tsv` (replaces `baseline_796495d_...`).

## What changed

| Change | Effect on Level 60 rotations |
|---|---|
| Mage: Ice Lance shares Fire Blast's category 19 cooldown; Fingers of Frost buff id 400670 -> 400669 in sim and the Frost/Frostfire APLs (d7541708d) | Frostfire -10.9 in baseline (published re-sim 1076.0 -> 1064.3). Frost unchanged (id rename only). Frostfire re-checked below: its order still wins |
| Warlock Bane of Doom hold, doomAlign 1 s (58a1c3593), all 12 warlock APLs | Spot-check, Affliction, old Doom line vs new, paired 20k, s0 1e6: +0.83 / +2.07 / +1.83 ±0.2 at 120/180/300 s. Matches the commit's +0.9 / +2.2 / +2.0 |
| Warrior Blood Craze (407f959a5) | A self-heal; no DPS effect on the DPS presets (no warrior row moved) |
| Restoration shaman healer APLs (raid_healer, tank_healer), sim/shaman/restoration | Healing sim only; not a Level 60 DPS preset |
| sim/core apl.go / apl_values_spell.go (healer values), raid_damage.go (Share, tank-only raids), tank_score.go, practice | No DPS preset moved more than 0.5 DPS except the mage/warlock rows above |

## Checks

- `apl_logic_diff.py 796495d6f a0b306984`: 16 files. Mage Frost/Frostfire FoF id swap (correct: 400669 is the buff the
  sim now registers, `sim/mage/talents.go:803`), the 12 warlock Doom lines, the two new resto healer APLs. No dead lines
  or wrong ids spotted.
- Baseline diff vs `baseline_796495d`: only Frostfire (-10.9) and the five warlock rows (+1.2 to +3.2) moved more than 0.5.
- Shared run vs published report: 0 of 73 moved more than 0.5 DPS (the re-sims already published).
- Open findings: none.

### Frostfire after the shared cooldown (paired 20k, B-A at 120 / 180 / 300 s; baseline A about 1065 at 180 s)

| Variant | s0 1e6 | s0 2e8 | Verdict |
|---|---|---|---|
| Drop the Fingers of Frost Ice Lance line | -27.9 / -69.7 / -64.2 | -28.2 / -70.2 / -63.9 | loses |
| Fire Blast lines above the FoF Ice Lance | -25.1 / -11.3 / -12.3 | -26.1 / -12.6 / -12.5 | loses |
| Drop Fire Blast (both lines) | -4.9 / +1.0 / -5.6 | -5.4 / +0.4 / -6.0 | fight-length trade-off, loses |

The current Frostfire order holds under the shared cooldown.

## Optional (under the bar)

Carried over, not re-measured: Mutilate window 0.5 s (probably moot), Fury-BS HS 55, hunter end-of-fight gates, Shadow
Mind Blast time-left gate, Destruction Fire DoT gates, Water Shield Magma gate.
