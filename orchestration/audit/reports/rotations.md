# Rotations pass 64 (2026-11-13, master 3a86d4a6b)

Last audited master cad23ae1a. What reaches the DPS raid presets:

| commit | change | reaches DPS presets |
|---|---|---|
| c7a99fa49 | Windfury Weapon's attacks (8232/8235/10486/16362) roll Windfury Totem under Forever | yes: every Enhancement build casts its own Windfury Totem with Windfury Weapon, +20 to +25 |
| 4c4790e3e | Flametongue Weapon rank 6 cut by 5.5 under Forever | no preset uses the imbue (Enhancement is on Windfury Weapon) |
| d22fe3e23 | Frightskull Shaft 3.9 PPM under Forever | no preset moved |
| 1743980e6 | Shadow Oil / Frost Oil procs keyed 350026 / 350025 under Forever | ID only; no preset moved |
| 041bb0a90 | Weakening Disease 15 over 14 s | no preset moved |
| tools/rotopt, enhancement bis.json | Phase 0 BiS reruns | not in the req files yet |

`apl_logic_diff.py cad23ae1a 3a86d4a6b`: 0 APL files touched.

## Baseline move (3000 it, seed 424242)

New baseline `ref/rotations/baseline_3a86d4a_3000it_seed424242.tsv` against 97da1cc (which was also cad23ae1a's state):

| build | old | new | diff |
|---|---|---|---|
| shaman__enhancement-level60 | 1030.8 | 1050.8 | +20.0 |
| shaman__enhancement-imp-gw | 1017.3 | 1037.6 | +20.3 |
| shaman__enhancement-mana-tide | 928.9 | 950.2 | +21.3 |
| shaman__enhancement-water-shield | 975.0 | 999.6 | +24.6 |

The other 64 presets are unchanged to 0.05 DPS. No `regress.py` run was needed: only the Enhancement
change reaches the presets, and it's attributed by the commit.

## Open item: Flametongue Weapon imbue with Windfury Totem

Checked on shaman__enhancement-level60, req `shamanImbue` WindfuryWeapon -> FlametongueWeapon, 20k it, 8 blocks,
S0 1e6, same seeds (two unpaired `ab.py` runs, one per req): 1064.9 / 1053.8 / 1037.2 vs 935.6 / 921.6 / 903.1
at 120 / 180 / 300 s, so FT loses about 132 at 180 s. Both commits in this range favour Windfury Weapon further.
Not a finding; the item can close for the level 60 builds. The Enhancement re-search in ada1d1e71's brief
covers the APL itself after the change.

## Open findings

All three still unapplied on master: rotations-20261112-1 (rogue poisons; presets.ts still Deadly MH),
rotations-20261109-1 (Fury 2H 90-rage Heroic Strike still at level60_fury_2h.apl.json), rotations-20261105-1
(feral Ferocious Bite gate still 45).

## Optional

None new.
