# Rotations pass 54 (no findings)

- Master: 0facd9428 (from 6b480b060, 335 commits), audited 2026-11-04. Client 1.60.1.70753 plus server hotfix
  push 113380 (shaman).
- Not quiet: sim `.go` changed (shaman, warrior, druid, paladin, priest, core). APL logic changed in 23 files,
  all tank or healer rotations (feral tank, Protection paladin, tank warrior, Warden, healers), out of this
  pass's DPS scope. No Level 60 DPS APL changed in logic.
- Regression baseline refreshed: `ref/rotations/baseline_0facd94_3000it_seed424242.tsv` (rows that moved are in
  `ref/rotations/MANIFEST.md`).

## What moved (baseline, 3k it, seed 424242, vs 6b480b0)

| Builds | Move | Cause |
|---|---|---|
| Feral cat, its FF duty | +26 | 74cc2b8e3: Natural Shapeshifter 2 -> Nature's Reach 2, hit-4 gear step. No powershift line, mana ends ~1600, no OOM. |
| Ret (level_60 and duties) | +10 to +12 | 3f4d44d4a: hit-3 gear step. Shock Ret unchanged. |
| Rogue (six) | +5 to +8 | 31c10b61c: hit+2/+3 gear steps. |
| Arcane | +7 | 74cc2b8e3: hit-1 step, Flame Throwing 1 -> Master of Elements 1. Mana ends ~1600, 0.7 s OOM. |
| Elemental, its Mana Tide | +7 | 87922bb77: hotfix 113380 (Lightning Bolt +2%, shocks, Chain Lightning ranks and mana). Near-uniform; LB/CL/shock order unaffected. |
| Enhancement Level 60, Water Shield, Mana Tide | +7.5 to +9 | 3e4796fe7: Improved Stormstrike 2 -> 1 / Tidal Focus 3, hit-3/-4 steps, plus hotfix shocks. Earth Shock gate keys on Stormstrike's time to ready, so it follows the talent. |
| Warrior | -2.9 to +1.3 | 9173d087a (Unbridled Wrath on autos only), 72c7d150f (Reckless Swings charges on landed swings), cd3f44cf2 (Fury DW hit+2). |
| Destruction fire | +0.8 | c8a7b4ac3: hit+1 step. |

`talent_check.py` passes every req (51 points, rows valid). `dead_lines.py` on the shared run shows no new dead
lines on the moved builds (only race, mana gem and talent-less lines, inert by design).

## Screens (`ab.py`, 20k it, 8 blocks, shared CLI 0facd9428f42)

| Preset | Variant | S0 | 120 | 180 | 300 |
|---|---|---|---|---|---|
| warrior arms_dragonkin (180 s base 1145.8) | Heroic Strike at >= 100 Rage (was 90), after Unbridled Wrath stopped proccing from HS | 1e6 | +1.03 ±0.84 | -0.63 ±0.65 | -0.47 ±0.32 |
| | | 2e8 | -0.25 ±0.72 | -1.52 ±0.61 | -1.67 ±0.54 |
| | | pooled | +0.39 | -1.08 | -1.07 |

The 90 Rage gate holds.

## Optional

None new. The pass 26 optional list (see `passes/rotations.md`) stands.

## Open items

- Feral and Enhancement were regeared onto new talents without APL re-screens here; the workers' own paired
  searches (`orchestration/round_70753/hit_trade/`) cover them and no line went dead.
- Machine load was ~18 on 4 cores during this pass, so only the one screen above was run.
- The rest of `passes/rotations.md` "Open items" is unchanged.
