# Results pass 101, master eceb090196 (2026-11-24)

Shared run (68 builds, 20k, seed 1, `/tmp/audit-work/runs/master`). Cause of the big change: the Phase 0 gear reset (step 4 regears of every class) plus the racial and regen fixes (7a2311b36, ed1df1d51, 9c08e16a2, e19d2df85, 0352d95dd).

| Check | Result |
|---|---|
| Run vs meta published.dps | 16 of 68 within 0.15, 32 off by more than 0.5 (metas not all refreshed) |
| Largest meta lags | Lone Wolf -22.4, rogues -15.1 (combat-swords), -14.0, -12.7, -5.9; mutilate +4.2, subtlety +3.2; hunters -1.0 to -4.6; rest <= 1.1 |
| Run vs builds.json | 53 of 68 differ by 1.5 or more; builds.json still holds Phase 1 numbers (warlock -160 to -210, fury 2H -140, Balance/mage/feral -110 to -135, Ret -69 to -100, Arms -64 to -92) |
| Drift report vs reference | only the known feral sapper line (results-20261105-1) |
| Unread fields | none |

Conclusion: no sim error found. The dashboard needs the full re-sim and republish for Phase 0, and the metas of hunters and rogues need refreshing. Ranking by the run: Lone Wolf 1291, Marksmanship 1200, BM 1113, Survival 1007 lead; warriors (Fury 1007, Arms 1016, Fury 2H 991) and rogue Combat 948 follow; Demonology 894, Fire 921, Arcane 843, Frost 727, Smite 695 are lowest of the casters. Not re-analysed in depth this pass.
Result: no new finding; update added to results-20261104-2.
