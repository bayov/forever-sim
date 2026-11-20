# Mechanics pass 93 (2026-11-20, master e07e8fdb8)

Range e671e881e..e07e8fdb8 (four sim commits, all item effects). Client unchanged at 1.60.1.70959.
No new findings. Open findings mechanics-20261118-2, 20261118-1, 20261113-2 and 20261104-1 stay open
(none touched by these commits).

## Get current

| Source | Newest | New since pass 92 |
|---|---|---|
| Client build | 1.60.1.70959 (`wago.py latest --probe 300`) | none |
| Dev notes t/2360696 | post 10, 2026-11-04 19:35Z | none, no edits |
| Hotfix post t/2403040 | post 12, 2026-11-18 03:24Z | none, no edits |
| Blue tracker | Kaivax, Forever Podcast ep. 8 (2026-11-19 17:35Z) | video only, no written changes; not reviewed |
| Daily update / TASKS | no TASKS.md, update_ledger.md or forever-update merges in range | none |

## Commits

| Commit | Change | Verdict |
|---|---|---|
| db684c04a | Bolt's Charged Eyeball 7.5% proc, 30% of shocks on the wearer | Departs from the client's 15% on a tier 1 log (BRS, 117 / 1518, 36 / 117). Accepted; verdicts row updated. |
| c04e51c13 | Blackhand's Breadth mark, Eye of the Beast, Weakness Analyzer, Adaptive Combat Assistant; core `CritChanceAgainst`, `MobFamily`, `OutdoorTerrain` | 1318944 aura 306 bp 5 (crit chance taken from the caster) matches. The new core term is added once in PhysicalCritChance (the only melee/ranged crit path, spell_outcome.go:797) and once in SpellCritChance; it is indexed by AllUnits, so player heal targets never collide with a marked enemy. Weakness Analyzer 1291101 aura 290 5% 1 charge, ACA 1291097 450 absorb 15 s match. The x2 on Orcs and Ogres is tooltip only (stated in code); no preset target sets a family, so no DPS effect. |
| 617a63635 | Item conditions registry (metadata for BiS ranking) | No mechanic. |
| 449bff490 | Forever values for BiS items | Totem of Rage 27859: aura 108 bp 2, misc 0 (direct) and misc 22 (DoT), shock mask; sim +2% additive on every shock hit and tick, flat 30 gone. Matches. Second Wind 63/s, Creeping Mold 42 x4, Seeping Willow 51 x5, Brain Hacker 250, Treant Slayer 175, Smokey's Lighter 210 all match 70959 SpellEffect. |

## Not checked

- Draconic Infused Emblem's 8% / 35 SP and the area bonuses (Rune and Jewel of the Guard Captain, Briarwood Reed): out of
  preset scope (outdoor conditions, not set by any Level 60 preset); rates are wowsims/forever log measurements.
