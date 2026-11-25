# Gear pass 77 (2026-11-25, phase0-step4-publish 01cc70b80)

Not quiet in range, no finding. Range 63129ce35..01cc70b80 plus the branch's fix commits: every DPS request on its
Phase 0 set (01cc70b80), 78802e6e's real on-use trinket lines, Feral's Maelstrom (02cf1139b), Fury DW's Tracker's
Signet (365c7c56f), Smite's set (347151b49), Smite Dwarf's hit (3ea3e4a1a), race_gear's Dwarf hit trade (f671944b3).

## Checked

| Check | Result |
|---|---|
| `phase0_check.py` | 3 `!` lines: Balance finger (Channeler's Ring), Shadow chest (Robes of Dark Summons), Shadow finger (Eye of Orgrimmar), each an open item 1 to 2 DPS above the pick at the list's iterations. The class notes (02cf1139b, 365c7c56f) re-ranked those slots at 20k and found them noise. Not findings. |
| Requests wear their sets | every DPS request's `gearSet` is its Phase 0 set; no `_hit` step left in the base requests |
| Hit (step 6) | final stats of the shared run: no build over its cap with a hit talent still taken and unexplained; Arms 8.3% (+0.3, Precision explained in the warrior step 3 notes), Survival 9% with Surefooted (re-checked in the hunter step 4 notes) |
| On-use trinkets the APLs cast (78802e6e) | Chained Essence of Eranikus (10 min) and Blackhand's Breadth (5 min) cast once a fight, Talisman of Ascendance 3.3 to 3.45 times; matches their cooldowns. Two lines lost DPS: rotations-20261125-1, fixed in 67fe28cf0 |
| List `dps` against the report (step 7) | not comparable: the list's `dps` is the search's score (uses valued on a 90 s fight, the user's talents, bis_phase0.sh header), so every build runs 2 to 20 below its list (Enhancement -2.6, Fury 2H -2.0, Affliction -20.2 before the fix). No list is stale against a sim change that moved its class except through 78802e6e and 1fed43cb, which only the rotation lines touch. |
| Enhancement imbue pairings (standards 40) | consumables report: Windfury Weapon 990.6 against Flametongue Weapon with Windfury Totem 883.2 on the Phase 0 set |

## Open findings

None open in `findings/gear.md`.
