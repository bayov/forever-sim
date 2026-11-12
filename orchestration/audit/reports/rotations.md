# Rotations pass 60 (2026-11-12, master a0b4d3be9)

Last audited master 71994fca6. 13 commits; those that reach the raid presets:

| commit | change | reaches raid presets |
|---|---|---|
| a0b4d3be9 | Windfury Totem rolls on main hand abilities under Forever (client flag list, `sim/core/windfury_totem_procs_forever.go`): Heroic Strike, Cleave, Sunder, Execute, Hamstring, Raptor Strike, Sinister Strike, Backstab, Mutilate... not Bloodthirst, Whirlwind, Stormstrike | yes (warriors, rogues, Survival) |
| e2322baa1 | Venom's +20% poison chance reaches Instant Poison | yes (rogues with Venom) |

`apl_logic_diff.py 71994fca6 a0b4d3be9`: 0 APL files. No req file changed.

## Baseline move (3000 it, seed 424242)

`regress.py` old CLI 71994fca6 vs new a0b4d3be9 on the same req files; the new baseline
(`ref/rotations/baseline_a0b4d3b_3000it_seed424242.tsv`) matches it:

| builds | DPS move |
|---|---|
| Arms (4 builds) | +168 to +171 |
| Fury 2H (3) | +98 to +106 |
| Fury DW (3) | +48 to +64 |
| Survival (3) | +69 to +72 |
| rogues (6) | +22 (Combat Daggers) to +58 (Seal Fate Swords) |
| everything else | under 1 |

`runs/vs_published.txt` compares against the published report, which predates the 2026-11-12 deploy
inputs; it isn't a sim-change table (warriors read -45 to -119 there but +48 to +171 old vs new CLI).

## Checks (paired `ab.py`, 20k it, 8 blocks, both seed sets unless noted; B-A pooled at 120 / 180 / 300 s)

| preset | variant | result |
|---|---|---|
| Arms | drop 90-rage Heroic Strike (finding 20261109-1) | -1.2 / -4.1 / -5.9: now loses |
| Arms | HS gate 75 / 60 / 50 | -5.1 / -19.4 / -31.5 at 180 |
| Fury 2H | drop 90-rage Heroic Strike | +0.9 / +2.0 / +0.4: still gains (finding narrowed) |
| Fury 2H | HS gate 75 / 60 / 50 | -6.8 / -17.8 / -31.5 at 180 |
| Fury DW | HS gate 60 -> 40 / 50 / 75 | -7.6 / -3.2 / -0.4 at 180: 60 holds |
| Survival | Raptor Strike above lines 5-6 | 0.00 (identical) |
| Survival | Wing Clip mana gate 50% -> 30% | +0.1 / +0.9 / +3.6: optional |
| rogues | MH Instant / OH Deadly (swap) | Mutilate +14.3, SF Daggers +16.3, SF Swords +9.8, Subtlety +2.4, Combat -2.7 / -2.8 at 180: **new finding** |
| rogues | Instant on both hands | -13 to -20 (SF Daggers -1.5) at 180, set 1 |

Old-CLI check: the poison swap already gained +7.4 (Mutilate) and +6.3 (SF Daggers) at 180 s on 71994fca6;
this master doubles it.

## Findings

- New: rotations-20261112-1 (high): Mutilate, both Seal Fate builds and Subtlety should carry Instant Poison
  on the main hand and Deadly on the off hand.
- Updated: rotations-20261109-1 narrowed to Fury 2H (+2.0 at 180); Arms keeps its Heroic Strike line.
- Still open: rotations-20261105-1 (Feral Ferocious Bite gate), not landed.

## Optional

- Survival Wing Clip mana gate 0.5 -> 0.3: +0.1 / +0.9 / +3.6 at 120 / 180 / 300 s, both sets (SE under 0.15).
  Under the bar at 180 s; a long-fight gain only.

## Tooling

- `tools/audit/rotations/ab.py` gained `--req-b OTHER.json`: B runs another req (settings A/B such as poisons
  or imbues), with `- -` as the APLs.

## Not done

- Hamstring and Sunder Armor weaving for Windfury Totem on warriors (Arms already has a 10-rage Hamstring filler);
  Fury DW Hamstring filler not tried.
- Rogue poison swap not run on the duty variants (none exist for rogues) or on the second seed set for Combat Swords.
