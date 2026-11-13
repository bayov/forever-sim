# Fairness pass 39 (2026-11-13, master 97da1cc)

Result: quiet (no new finding). 38 watched files changed since 71562a8: the Ret re-tunes after the Nov 11 Judgement
hotfix (ec8ab9b3d), Unsealed Wrath and Volley hotfixes, Windfury Totem not rolled by extra attacks (a1a1a81bb, standards
17 amended), the item database (Nov 11-12 reveals, 97da1cca5), spell icons. Full race sweep (5k, seed 1, every base
request at every race, the loop's 97da1cc CLI and tree). Client 1.60.1.70938 (client_update pass 13: no shared change).

## What moved against the 71562a8 baseline (more than 1.5)

| Build | Before | Now | Cause |
|---|---|---|---|
| Retribution (Undead) | 1004.6 | 983.8 (-2.1%) | ec8ab9b3d Ret re-tune + Unsealed Wrath hotfix + extra attacks no longer roll Windfury Totem; Ret 15th -> 16th, behind Destruction Fire |
| Sub-Holy Ret | 993.5 | 974.3 | same |
| Shock Ret | 963.9 | 970.4 | ec8ab9b3d (Seal of Righteousness, Holy Shock first) |
| Rogue Combat Swords | 946.1 | 943.9 (-2.2) | a1a1a81bb: Sword Specialization swings no longer roll Windfury Totem (ui/rogue/presets.ts note: 943.8 at 20k) |

Everything else within 1.0 (Survival +0.9, Subtlety -0.2: noise). Warriors, Enhancement, hunters (Volley is off their
rotations) did not move. The extra-attack rule lives in sim/core (`AutoAttacks.IsExtraAttack()`), so it applies to every
class alike; only builds with extra-attack sources under the totem moved. No shared leak.

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS" (item DB regen changed no synthetic item).
- Best race: report race best or within noise everywhere (MM Night Elf +0.3, Fire Gnome +0.4, Combat Daggers Gnome
  +0.6; MM pet variants Night Elf +1.0 / +1.5). Paladins: Undead best (Human -4.5, Dwarf -11.4).
- Published: f6b04b9e7 re-simmed the Ret, rogue and Prot moves; the paladin metas now match (Ret +0.2, Shock +1.1,
  Sub-Holy +0.4). Other metas still lag (Fury 2H -119, Fury -86, Arms -46, Affliction +37, Survival -26, hunters,
  rogues -10 to -20, Feral +15): results-20261104-1 / -2 cover them; builds.json matches the 20k run (results pass).
  Confirmed independently here, not refiled.

## Ranking now (report race, 5k)

Lone Wolf 1291.8, MM 1234.7, BM 1182.0, Feral 1153.8, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0, Enhancement 1030.7, Survival 1014.8,
Destruction Fire 997.9, Ret 983.8, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6, Combat Swords 943.9,
Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7, Subtlety 863.4, Smite 781.0.
Top over bottom 1.65x.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this full sweep. Baseline sha 97da1cc.

## Open for the next pass

- Metas' published.dps: re-check when results-20261104-1/-2 are fixed.
