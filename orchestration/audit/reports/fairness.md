# Fairness pass 40 (2026-11-13, master 3a86d4a)

Result: quiet (no new finding). Watched files changed since cad23ae1: sim/core/buffs.go (Windfury Weapon's attacks roll
Windfury Totem under Forever, c7a99fa49, our Nov 13 log 47/162), sim/shaman/flametongue_weapon.go (rank 6 cut 5.5,
4c4790e3e), sim/common/item_effects.go (Frightskull Shaft 3.9 PPM, disease 15 x 7, d22fe3e23 / 041bb0a90),
sim/core/consumes.go (Shadow / Frost Oil procs keyed 350026 / 350025, 1743980e6, IDs only). Full race sweep (5k, seed 1,
every base request at every race, the loop's 3a86d4a CLI and tree). Client 1.60.1.70938.

## What moved against the 97da1cc baseline

| Build | Before | Now | Cause |
|---|---|---|---|
| Enhancement (Orc) | 1030.7 | 1054.1 (+2.3%) | c7a99fa49 Windfury Weapon rolls the totem, net of the rank 6 Flametongue cut; 13th -> 10th, ahead of Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0 |
| Enhancement imp-gw | 1017.3 | 1038.3 | same |
| Enhancement mana-tide | 931.1 | 951.4 (Orc; Windshaper 951.5, tie) | same |
| Enhancement water-shield | 977.0 | 1000.1 | same |

No other line of the sweep changed at all (diff of sweep logs touches only these four). The buffs.go change is keyed on
the imbue's own spell IDs, so it can't reach another class; the oil ID change and Frightskull touch no report request.
No shared leak. The rule is class-specific and log-backed (our log tier), so it isn't a fairness question.

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Best race: unchanged from pass 39; Enhancement still Orc (Windshaper ties within 0.1 on mana-tide).
- Published: the four Enhancement metas lag the sweep by +22 to +23.5. Results raised it this run under
  results-20261104-2 (update 2026-11-13); confirmed independently here, not refiled. The re-sim is already briefed
  (orchestration/round_notes_2026-11-11/briefs/enh_1113.md, ada1d1e71). Other metas' gaps unchanged from pass 39
  (Fury 2H -119, Fury -86, Arms -46, Affliction +37, Survival -26, hunters, rogues), same results findings.

## Ranking now (report race, 5k)

Lone Wolf 1291.8, MM 1234.7, BM 1182.0, Feral 1153.8, Affliction 1071.1, Frostfire 1065.7, Destruction 1063.2,
Demonology 1055.3, Fire 1053.8, Enhancement 1054.1, Fury DW 1050.2, Arms 1041.7, Fury 2H 1032.0, Survival 1014.8,
Destruction Fire 997.9, Ret 983.8, Elemental 977.4, Arcane 963.6, Balance 952.1, Shadow 948.6, Combat Swords 943.9,
Frost 943.6, SF Swords 938.7, Combat Daggers 936.0, SF Daggers 904.4, Mutilate 865.7, Subtlety 863.4, Smite 781.0.
(Table order is by best race: Fire's Gnome 1054.2 sits just above Enhancement.) Top over bottom 1.65x.

## Baseline

`ref/fairness/sweep_baseline.log` and `table_baseline.md` replaced with this sweep. Baseline sha 3a86d4a.

## Open for the next pass

- Enhancement metas: check they land within 1.5 of 1054.1 / 1038.3 / 951.4 / 1000.1 after the enh_1113 re-sim.
- Other metas' published.dps: re-check when results-20261104-1/-2 are fixed.
