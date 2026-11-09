# Rotations pass 57 (2026-11-09, master 66a243cf8)

Last audited master 9ba00c2dc. 16 commits; three touch watched paths:

| commit | change | reaches raid presets |
|---|---|---|
| 0ce9dbf40 | Windfury Totem procs from white main-hand swings only under Forever (user ruling 2026-11-08) | yes |
| 3b0983bc6 | Orc shaman base attributes, levels 2-43 | no (leveling) |
| 3a99cf650 | two report assumptions confirmed (assumptions.json) | no |

`apl_logic_diff.py 9ba00c2dc 66a243cf8`: 0 APL files touched. Req files unchanged.

## Baseline move (3000 it, seed 424242, same reqs; all from 0ce9dbf40)

New baseline `ref/rotations/baseline_66a243c_3000it_seed424242.tsv` (replaces `baseline_9ba00c2_...`).
31 of 68 presets moved; every other one is within 0.05.

| group | DPS change at 180 s |
|---|---|
| Arms (4 builds) | -195 to -199 (-18%) |
| Fury 2H (3) | -179 to -189 (-16%) |
| Fury DW (3) | -100 to -109 (-9%) |
| Survival hunter (3) | -94 to -97 (-9%) |
| Ret paladin, Level 60 and Sub Holy (7) | -85 to -95 (-8%) |
| Ret paladin, Shock (3) | -58 to -59 (-6%) |
| Rogue swords (2) | -45 to -50 |
| Rogue daggers, Mutilate, Subtlety (4) | -20 to -29 |
| Feral cat (2) | -27 to -30 |
| Enhancement (4) | 0 (Windfury Weapon blocks the totem) |

Sanity check on Arms (2k it, 180 s, old vs new CLI): WF Totem extra attacks 27.6 -> 9.2 a fight, white swings
44.0 -> 47.6, Heroic Strike 8.3 -> 2.0 casts (less rage). 9.2 is about 20% of the white swings, as the new
proc mask says. The size matches the ruling; nothing looks broken.

## Screens (paired ab.py, 20k it, CLI 66a243cf8a64, B-A at 120 / 180 / 300 s)

| preset | variant | S0 1e6 | S0 2e8 | verdict |
|---|---|---|---|---|
| arms | delete HS 90 line | +2.36 / +3.48 / +2.83 | +2.18 / +3.49 / +2.09 | **finding 20261109-1** |
| arms | HS gate 90 -> 100 | +1.29 / +1.59 / +1.19 | +0.74 / +1.17 / +0.66 | dominated by delete |
| arms | delete Slam | -58.4 / -57.9 / -57.2 | | keep Slam |
| fury_2h | delete HS 90 line | +2.19 / +2.88 / +1.58 | +1.43 / +4.13 / +2.12 | **finding 20261109-1** |
| fury_2h | HS gate 90 -> 100 | +0.96 / +1.09 / +1.27 | +0.87 / +1.51 / +1.66 | dominated by delete |
| fury_2h | delete Slam | +0.88 / -0.04 / -0.34 | | flat, keep |
| fury (DW) | HS gate 60 -> 80 | -4.81 / -2.81 / -3.59 | | keep 60 |
| fury (DW) | HS gate 60 -> 50 | -2.66 / -1.69 / -1.25 | | keep 60 |
| survival | delete Raptor Strike | -37.1 / -35.8 / -33.0 | | keep |
| enh mana-tide | Flametongue Weapon (req) | -28.3 / -28.5 / -30.1 | -27.9 / -28.4 / -30.6 | 20261108-1 withdrawn |
| enh water-shield | Flametongue Weapon (req) | -31.9 / -32.2 / -34.3 | -31.9 / -32.3 / -33.9 | 20261108-1 withdrawn |

## Findings

- New: **rotations-20261109-1** (low): drop the 90-rage Heroic Strike line from the Arms and Fury 2H APLs, +3.5
  at 180 s pooled on both.
- Withdrawn: rotations-20261108-1 (Flametongue Weapon for Mana Tide / Water Shield). With Stormstrike no
  longer procing the totem, it flips to -28 / -32. It was never applied, so the presets need no change.
- Still open: rotations-20261105-1 (Feral Ferocious Bite gate 45 -> 65); master still has 45.

## Optional

None new.

## Open items

- Ret paladin lost 58-95 (Seal of Command and Crusader Strike hits no longer proc the totem). The twist rule
  and seal choice weren't re-screened this pass.
- Rogue (Sinister Strike / Backstab no longer proc it): no rotation knob obviously depends on it; not screened.
- Arms/Fury 2H duty and Dragonkin builds weren't run separately for 20261109-1.
- The ruling leaves out on-next-swing attacks (Heroic Strike, Raptor Strike) unchecked in the logs; if logs show
  they do proc the totem, 20261109-1 should be re-measured.

## Tooling

`tools/audit/rotations/ab.py` takes `--breq B_REQ.json` for settings A/B (e.g. an imbue) with `- -` as the APLs.
