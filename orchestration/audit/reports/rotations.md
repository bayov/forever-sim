# Rotations pass 51 (no findings)

- Master: f331f0cff (from f23ea25d2), audited 2026-11-04. Client 1.60.1.70753 / 70755 (70755 identical tables).
- Not quiet by the diff rule: 6 Level 60 APLs changed in logic, and non-test sim `.go` changed.
- New regression baseline: `ref/rotations/baseline_f331f0c_3000it_seed424242.tsv` (old f5c99ad one removed).

## What changed

| Commit | Change |
|---|---|
| 4f8edc219 | Feral (`level60`, `level60_ff`): EZ-Thro Dark Bomb (item 260817) on its own line at <=60 Energy right after a main-hand swing; Rake and Shadowmeld Ravage also at 5 combo points. Balance (`level60`, `level60_ff`): Clearcasting Starfire line dropped, Wrath (9912) fallback after the Starfire lines. |
| e799dccfa | Ret twist and sub-Holy: Holy Strike (10333) moved ahead of Judgement and the seals. |
| da726bbe1 | Priest Smite preset: Spiritual Guidance talents (user-approved, round 70753 priest.md). |
| 13c34f20f | Sim: notes-only values gated under Forever (rogue, hunter, paladin Vengeance, shaman, warlock demon health). Outside Forever only. |
| 0dcb0d87a, 90e0bb01a | Sim/settings: tank threat (Bear Form, Defensive Stance, Defiance); Salvation removed from raid buffs. Threat only. |

## Regression check (`baseline.py`, 3,000 it, seed 424242, shared CLI f331f0cff0d8)

Against `baseline_f5c99ad` every row within 0.05 DPS except:

| Preset | old | new | moved | cause |
|---|---|---|---|---|
| druid__feral-cat | 1130.52 | 1138.73 | +8.21 | bomb line |
| druid__feral-cat-duty-ff | 1127.80 | 1134.09 | +6.29 | bomb line |
| druid__balance | 952.49 | 953.59 | +1.10 | Wrath fallback |
| druid__balance-duty-ff | 915.79 | 916.13 | +0.34 | Wrath fallback |
| paladin__level_60 | 1090.09 | 1090.63 | +0.54 | Holy Strike first |
| paladin__level_60_sub_holy | 1090.77 | 1092.10 | +1.33 | Holy Strike first |
| priest__smite | 762.14 | 781.25 | +19.11 | talents |
| warlock__affliction (+2 duty) | 1016.08 | 1035.19 | +18..+19 | Wrack build, 734fe0e93 (pass 50) |
| shaman__enhancement-level60 | 1158.50 | 1158.37 | -0.13 | RNG (Salvation line dropped) |

So the sim changes move no Forever DPS, as 13c34f20f claims; `regress.py` was not needed.

## Verification of the shipped rotation changes (paired, `ab.py`, 20k as 8 x 2,500, old req rotation vs new)

| Preset | seed set | 120 | 180 | 300 |
|---|---|---|---|---|
| Feral (base 1131.1) | 1e6 | +8.01 ±0.38 | +7.67 ±0.41 | +7.35 ±0.30 |
| Feral | 2e8 | +7.70 ±0.43 | +8.09 ±0.39 | +7.53 ±0.22 |
| Balance (base 952.2) | 1e6 | +0.73 ±0.18 | +0.39 ±0.18 | +1.69 ±0.16 |
| Balance | 2e8 | +1.02 ±0.16 | +0.61 ±0.17 | +2.07 ±0.11 |

Both hold on both seed sets and match the druid worker's numbers (round_70753/druid.md). Balance is under
the 1 DPS bar at 180 s but positive everywhere, so shipping it is fine. Paladin's change was measured by its
worker on two seed sets (+1.2/+1.3 twist, +1.4/+2.2 sub-Holy at 180 s); the baseline agrees in sign; not re-run.

## Notes for others

- `runs/vs_published.txt`: the six shaman presets are 26-33% above their `published` meta values (e.g.
  Enhancement 883.5 published vs 1159.4). Master matches pass 49's baseline, so this is stale published
  data from before the shaman rework, not a rotation or sim regression (results auditor's area).
- Priest Smite (Spiritual Guidance) drops to 713.5 at 300 s against 781 at 180 s (mana). The worker's knob
  search found only fight-length trade-offs; not re-screened here.

## Optional

- None new.

## Open items carried

- Balance Night Elf Elune's Light vs Tauren, Elemental Mastery hold, warlock duty-curse variants (see pass doc).
