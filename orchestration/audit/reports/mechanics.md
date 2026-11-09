# Mechanics pass 71 (2026-11-09, master 66a243cf8)

No new finding. Both open findings (mechanics-20261107-1 natural flasks, mechanics-20261104-1 druid scales) are still open: no commit touched them.

## Range
9ba00c2dc..66a243cf8, two sim commits (no merges from forever-update, TASKS.md only a test-failure note).

| Commit | Change | Verdict |
|---|---|---|
| 0ce9dbf40 | Windfury Totem procs only from white main hand swings under Forever (`ProcMaskMeleeMHAuto`, sim/core/buffs.go WindfuryTotemBuffAura) | User ruling 2026-11-08 "for now", backed by beta logs (tier 1): 882 totem procs, none from about 2,770 Stormstrikes where 20% would give about 550. Not a finding. Open item: on-next-swing attacks (Heroic Strike, Cleave) are excluded too and unchecked in logs; the commit says so. A warrior log with the totem up would settle it. Recorded in verdicts.md. |
| 3b0983bc6 | Orc shaman base attributes at levels 2-43 from BayovCore's character sheet (sim/core/base_stats_levels_auto_gen.go) | Leveling only; levels 1 and 44-60 unchanged, so no Level 60 preset moves. Source is live character sheet (tier 1). Not a finding. |

## Get current
- Client: newest 1.60 build still 1.60.1.70862 (probe 300).
- Dev notes thread t/2360696: 8 posts, newest post 10 (2026-11-04T19:35:16Z), post 1 edit 2026-11-03T03:20:46Z; unchanged.
- Blue tracker since 2026-11-07T01:07:59Z: one Forever post, "Dungeon Enemy XP Increased - November 8" (t/2406595, Kaivax 2026-11-08T19:47:33Z): "increased kill experience in dungeons by 50%". Not a combat mechanic; relevant only to leveling XP models.
