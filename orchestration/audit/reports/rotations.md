# Rotations pass 39 (2026-10-30)

- **Master:** d50909a37 (last audited 334c7decd). Client 1.60.1.70630 data, unchanged.
- **Result:** quiet, no new findings. rotations-20261028-1 (hunter Trueshot/Hawk caveats) still open, now
  also in the four new hunter pet variants' meta files.

## What changed

64 commits, nearly all the raid sim (sim/compsearch, comp resolver, per-player raid buffs, camp buffs,
multi-player raid bug fixes) and its 19 new report variants. `apl_logic_diff.py 334c7decd d50909a37`:
- New duty APLs: paladin Seal of Wisdom / Crusader duties (4), warlock Curse of Recklessness duties (4),
  warrior Sunder duties (3). New reqs also add hunter Tallstrider and Wolf pet variants (6; rotations
  unchanged from their bases).
- 0cf1a8c8b Feral `level60_ff`: Faerie Fire (Feral) 17392 replaced by caster Faerie Fire 9907 (free in Cat
  Form, 1.5 s GCD, no swing reset; Forever has no Feral FF). Duty cost at 180 s rises 4.5 -> 8.3.

## Checks

- Shared 20k run vs published: only mage__frostfire moved (+0.9, the pass-38 Evocation change, not yet
  republished). The sim `.go` changes move no existing preset.
- Regression baseline (3000 it, seed 424242) vs baseline_1274e15: feral-cat-duty-ff -3.2 (intended, above),
  mage__frostfire +2.1, elemental +0.5/+0.2, shock_undead -0.3 (pass-38 APL changes). Everything else
  identical. New baseline: `ref/rotations/baseline_d50909a_3000it_seed424242.tsv` (adds the 19 variants).
- Duty upkeep, 300 it seed 7, uptime of the duty debuff on the target out of 180.9 s:

| preset | debuff | uptime s | casts/fight |
|---|---|---:|---:|
| warrior arms / fury / fury_2h -duty-sunder | Sunder 11597 | 180.6 / 180.8 / 180.9 | 11.7 / 11.2 / 10.9 |
| warlock aff / demo / destro / destro_fire -duty-recklessness | CoR 11717 | 180.5-180.6 | 2.0 |
| paladin 4 x -duty-wisdom | JoW 20355 | 180.9 | 1 (melee refreshes) |
| paladin 2 x -duty-crusader | JoC 20303 | 180.9 | 1 |
| druid__feral-cat-duty-ff | FF 9907 | 180.8 | 6.4 |

  No duty lapses.
- Feral FF placement A/B (FF only below 30 Energy unless < 2 s left), 20k, two seed sets, B-A pooled:
  120 s -0.82, 180 s -0.72, 300 s -0.37. Loses; the 7 s refresh holds.

## Not covered

- The new duty APLs' damage lines are their base rotations'; not separately A/B'd beyond upkeep.
- Older open items unchanged (Balance Night Elf Elune's Light, Elemental Mastery hold, warlock duty Life
  Tap/Doom screens).

## Optional

- None new.
