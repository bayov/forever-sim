# Fairness pass 63 (2026-11-26, master 2af970d74)

Full pass: range ab3fff700..2af970d74 is the Phase 0 prices round 2 (over-the-gold-line picks taken out and every
class's trinkets, weapons and ranged re-ranked), the rogue re-rank on the current sim, a Barov servant swing-wait
guard, mksettings explosives for Survival / Enhancement / Feral, the report re-sim of the re-ranked builds (aeaae077c,
db7482a0f) with published DPS refreshed (31d2f59e8), and warrior leveling presets. 5k race sweep (seed 1, every race,
the report's own requests with only the race changed) on the 2af970d74 CLI, table.py against the ab3fff700 baseline,
phase0_check.py, the loop's 20k run against published, and paired 20k item and race checks.

## What moved, and why

Every build that moved had its own request changed in range; every other build is exactly 0.0 (same seed). No leak.

| Build | Before | Now | Cause |
|---|---|---|---|
| Survival (and wolf / tallstrider) | 963.5 | 950.5 (-1.3%) | 13592ab69 trinket re-ranked without over-line picks |
| Fury DW | 994.6 | 987.5 (-0.7%) | ef9dbff7a trinkets and ranged under the 30g line |
| Enhancement (and variants) | 971.4 | 966.3 (-0.5%) | 518f23165 Heart of Wyrmthalak out |
| Lone Wolf | 1140.5 | 1136.2 (-0.4%) | 13592ab69 ranged re-ranked |
| Arms | 1012.9 | 1010.6 (-0.2%) | dac79e5f6 trinket back to Maelstrom |
| Feral | 1060.5 | 1058.6 (-0.2%) | f4e2b6e46 trinket back to Maelstrom |
| Shock Ret / Sub Holy | 904.8 / 905.2 | 902.0 / 908.1 (best race) | 518f23165 (Blackblade out) |
| Ret | 931.1 | 932.7 (+0.2%) | 518f23165 Destiny, Blackhand's Breadth |
| Mutilate, Seal Fate Daggers, Combat Daggers, Subtlety | | +2.6, +2.0, +0.7, +0.7 | rogue re-rank (e9da0ddcd, f3bf14c7c, 26e4300f1) |
| Balance | 802.4 | 801.9 | f4e2b6e46 off hand (noise) |

The only sim change for the report, the Barov servant guard (57fbb3401), only stops a panic when a swing is due now.

## Published against current

Every base build and variant is within 1.2 of its published number at its report race (31d2f59e8 refreshed them).
The earlier lags (Ret +7.5, Elemental +3.9, Affliction +3.4, Affliction Shadow Bolt +26) are gone; results-20261104-2
is the results auditor's to close.

## New finding

- **fairness-20261126-4 (medium):** Ret leaves out Darkmoon Card: Maelstrom as "doubtful", against standards 42 and
  unlike Arms, Feral, Fury DW and Seal Fate Swords, which wear it. Paired 20k: paladin__level_60 +3.83 for Eranikus
  (+3.42 for Blackhand's Breadth), Sub Holy +2.52. The round notes' 0.4 / 0.9 was measured on the old Blackblade set.

## Gear (standards 43)

- phase0_check.py: every DPS request wears its spec's set. Every list is now on 34fe148a5 or later (the rogues on
  6a5cacef8, which has the batch rules): fairness-20261126-1 resolved.
- Close calls the round notes measured and kept: Destruction Spellblade (+1.75 in the list), Enhancement Blackrock
  Slicer (+1.77 in the list, 1.5 / 0.6 at 20k in the notes). Ret's Maelstrom is the one that doesn't hold (above).

## Open findings, rechecked

- fairness-20261126-3 (Shahram stun): no report request wears Blackblade of Shahram now, so no report number is
  affected; the code bug remains. Kept open, low.
- fairness-20261125-2 (explosives): Fury, Fury 2H, Mutilate, both Combat, Seal Fate Swords and the Ret builds still
  carry no filler explosive. Not re-measured.
- Race, 5k: Demonology Undead +4.2, MM High Order +5.1 (variants +6.6), Balance Wrath High Order +4.0, Arcane Undead
  +1.9, Destruction Undead +1.7, Fury DW Undead +1.2, Fire Gnome +1.1. New, below a finding: Sub Holy Dwarf over Undead
  +1.65 paired 20k (905.3 / 905.1 -> 907.2 / 906.5; Dwarf is Alliance, the set is Horde), Shock Ret Dwarf +1.1 at 5k.

## Ranking

Top over bottom 1.67x (Lone Wolf 1136.2 over Smite 679.5). Survival drops from 9th to 10th, behind Affliction; no other
place changes. Full table in orchestration/fairness_report.md §1.

## Baseline

sweep_baseline.log and table_baseline.md replaced with this sweep (2af970d74).
