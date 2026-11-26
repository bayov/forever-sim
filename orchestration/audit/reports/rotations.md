# Rotations pass 80 (2026-11-25, master 85691655a)

Result: findings (1 new, low). Range c3bbe33ea..85691655a (21 commits): the wowsims/forever evidence batches (tank
threat, caster resists and overloads, hunter pet rules, item procs, proc rules) and the 2026-11-25 ruling batch
(22d2cf952: EV-0088 explosive throws keep the swing timer and stun explosives miss level 63 bosses; EV-0008 extra
attacks roll Windfury Totem; EV-0081 demon melee crit).

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py c3bbe33ea origin/master` | 0 APL files touched |
| `tools/dps_report/req/*.json` changed | none (the hunter/feral explosive move landed only in presets.ts and mksettings; filed by consumables-20261125-3 and results-20261125-1) |
| Non-test sim `.go` changed | 68 files, all EV-* batches above; reached by every class |
| Open findings landed? | no: rotations-20261122-1 and rotations-20261112-1 stay open |
| Regression baseline (68 presets, 3000 it, seed 424242) | refreshed: `ref/rotations/baseline_8569165_3000it_seed424242.tsv` replaces the 0404484 one |
| Resource check (shared 20k run, `secondsOomAvg`) | unchanged: worst warlock Affliction SB 4.5 s, other warlocks 3.1-3.4 s, Enhancement 2.5 s, all fine |
| EV-0088 swing-timer gates on explosive lines (rogue, feral) | measured below; feral filed as rotations-20261125-1 |

### Baseline moves vs 0404484 (3000 it, seed 424242)

Hunters +3.0 to +4.9% (pet white swing takes the hunter's crit, EV-0134); rogues -1.4 to -3.0% (Mutilate -3.0%; weapon
PPMs and set procs, EV-0144); Fury duty Battle Shout -1.5%; other builds within about 1-2%. Sim mechanic changes, not
rotation faults.

### Explosive swing gates after EV-0088 (paired, 20k it, 8 blocks, B-A pooled over seed sets 1e6 and 2e8)

| Build | Variant | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| rogue__subtlety | Land Mine without the swing gate | +0.47 | +0.44 | +0.41 |
| rogue__seal-fate-daggers | Land Mine without the swing gate | +0.14 | +0.10 | +0.50 |
| druid__feral-cat (req with Land Mine) | 260817 -> 4395, energy gate only | +0.49 | +0.58 | -0.01 |
| druid__feral-cat (req with Land Mine) | 260817 -> 4395, both gates | +0.41 | +0.52 | -0.02 |
| druid__feral-cat-duty-ff (req with Land Mine) | 260817 -> 4395, energy gate only | +0.62 | +0.52 | -0.04 |

Per-set SEs 0.1-0.3. All under the 1 DPS bar; the feral line is filed for its wrong item id (correctness).

## Open findings

New: rotations-20261125-1 (feral APLs name the Dark Bomb the presets dropped). Unchanged: rotations-20261122-1,
rotations-20261112-1.

## Optional

- Rogue Land Mine lines (all six rogue APLs): drop the `autoTimeToNext >= autoSwingTime - 0.4s` term and the stale
  "1 sec cast loses no swing time" note; +0.4 Subtlety, +0.1 Seal Fate Daggers at 180 s (above). Mostly a note fix.

## Open items

Not done this pass (bounded by time and a loaded machine): the Phase 0 trinket line check, the Windfury main-hand timing
tests, and rotation re-checks for the new caster rules (Maelstrom Ready! keeping Clearcasting for Enhancement/Elemental
Chain Lightning, EV-0180; judgements refreshing only on white swings and totem hits for paladins, EV-0117; Revelation
spent on cast completion, EV-0028). The last three moved builds 1-2% and are worth a fan-out pass.
