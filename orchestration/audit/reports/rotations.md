# Rotations pass 53 (no findings)

- Master: 6b480b060 (from fb584eafa, 97 commits), audited 2026-11-04. Client 1.60.1.70753 plus server hotfix
  pushes 113483 (rogue) and 113539 (warrior).
- Not quiet: sim `.go` changed for rogue, warrior, shaman and core, and 24 APL files changed in logic.
- Regression baseline refreshed: `ref/rotations/baseline_6b480b0_3000it_seed424242.tsv` (rows that moved are in
  `ref/rotations/MANIFEST.md`; only rogue and warrior moved).

## What changed

| Commit | Change | Rotation effect |
|---|---|---|
| 316a600ac | Rogue hotfix 113483: Shank real IDs (1329549), always hits, no poison / enchant / Hack and Slash procs; Mutilate flat 44/63/88, casts 1241584; Venom chance on Deadly and Wound only | 7 rogue APLs swap Shank 1399248 -> 1329549 and Mutilate 1310707 -> 1241584 (IDs match the commit). Seal Fate -16 to -18, Mutilate -8 at 3k it. Published numbers are stale (vs_published: 5 rogue rows), report not regenerated yet. |
| 7cd399e88, e07adf11b | Warrior hotfix 113539: Raging Blows out, Reckless Swings 1329555 on a 1 s GCD, Spearing Strike 75% | 10 warrior APLs: Reckless Swings moved below the strikes after the sunder sequence; 2H Fury, its duties, Arms and Arms Battle Shout cast it only under 20 Rage. Talents re-picked. |
| 8f304b240, b57f5cc0a | Downranking penalty in sim/core, now on damage too | No Level 60 row moved (top ranks; Smite rank 7 learn level 54 is inside the 17-level margin). |
| 280fd26e2 | Chain Lightning coefficient 0.714 at every rank | No row moved: no Level 60 rotation casts it on one target. |
| healers | Holy paladin, Holy / Discipline priest APLs and the absorb pipeline | Healer presets, out of this pass's DPS scope. |
| ca0ca1285 | dps_report sims with labeled random streams | Report tooling only. |

## Screens (`ab.py`, 20k it, 8 blocks, shared CLI 6b480b06046c), B-A pooled over S0 = 1e6 and 2e8

| Preset | Variant | 120 | 180 | 300 |
|---|---|---|---|---|
| rogue seal-fate-daggers (180 s base 904.0) | Shank at <= 60 Energy (was 75) | -0.54 | -0.49 | -0.58 |
| | Shank at <= 90 Energy | +0.09 | -0.03 | +0.03 |
| | no Shank | -98.5 | -99.2 | -100.0 |
| rogue combat-swords (959.0) | Shank at <= 60 Energy | +0.08 | +0.08 | -0.03 |
| | Shank at <= 90 Energy | -7.61 | -5.25 | -2.63 |
| | no Shank | -159.9 | -160.8 | -158.3 |
| shaman elemental-level60 (969.8) | Chain Lightning 10605 while Elemental Mastery (16166) is up, above Lightning Bolt | -15.7 | -18.6 | -74.6 |
| | same, only at >= 40% mana | -15.7 | -14.7 | -26.3 |
| warrior arms-duty-sunder (1040.9) | Reckless Swings only under 20 Rage (added) | -1.61 | -1.61 | -1.94 |
| warrior fury, dual wield (1136.0) | Reckless Swings only under 20 Rage (added) | -11.0 | -6.99 | -7.08 |
| warrior arms (1088.4) | Reckless Swings without the 20 Rage gate | -1.44 | -2.40 | -1.56 |

Per-set SE 0.1 to 1.2; the two seed sets agreed on every sign. Shank at 75 Energy still holds after the hotfix,
the worker's per-build Reckless Swings gates hold, and Chain Lightning stays out of Elemental even under
Elemental Mastery (mana, as the shaman wiki says).

## Optional

None new. The pass 26 optional list (see `passes/rotations.md`) stands.

## Open items

- Rogue: the hotfix moved Shank's value (no procs); only its Energy gate was re-screened here. Combat Daggers,
  Mutilate and Subtlety were not screened separately.
- The rest of `passes/rotations.md` "Open items" is unchanged.
