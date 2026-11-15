# Mechanics pass 83 (2026-11-15, master 70ea588bb)

Quiet: no new finding.

## Get current

- Client: 1.60.1.70938 still the newest (`wago.py latest --probe 300`).
- Blizzard: dev notes t/2360696 unchanged (post 10, 2026-11-04; post 1 edited 2026-11-03). Hotfix thread
  t/2403040 unchanged since pass 82 (post 10 and post 1 edit, 2026-11-14 01:54 UTC). Blue tracker: one newer post,
  "Player-Level Dungeon XP Adjustment - November 14" (2026-11-15 01:18 UTC), XP only, no DPS mechanic.
- Daily update: ad64ef0f8 (TASKS) records the user ruling of 2026-11-15 to keep Windfury Totem's AP after hotfix push
  114091 (the in-game tooltip still shows 246 AP) until a log settles it. A ruling, not a finding; the sim keeps the AP.

## Commits d630df2cb..70ea588bb

| Commit | Change | Verdict |
|---|---|---|
| 0cc2c3c8b | Ferocious Bite: a missed / dodged / parried Bite refunds 80% of its 35 Energy under Forever, keeps extra Energy and combo points | Verified tier 1: Attributes_1 0x8000000 (DISCOUNT_POWER_ON_MISS) on every player rank (22568, 22827-22829, 31018, 1328618) in 70938, same as the builders and rogue finishers the sim already refunds. `EnergyCost.IssueRefund` refunds 0.8 x CurCast.Cost (0 under Omen of Clarity); extra Energy is spent only on a landed Bite. Net 7 Energy a miss fits the 6.1-6.7 measured on a 70862 log (regen in between). |
| 0cc2c3c8b | Sharpened Claws adds spell crit in feral forms | Verified tier 1: 16942 effect 6 aura 290 (all crit). About 0 DPS. |

Pounce (9005/9823/9827, 1229228/1258520) carries the same flag but has no refund in the sim; no Level 60 build uses it.

## Open findings

Unchanged: mechanics-20261114-1, -20261113-2, -20261113-1, -20261112-1, -20261104-1.
