# Consumables pass 42 (master 74b3a5f7c, client 1.60.1.70532)

Result: quiet. No new findings, nothing open.

## Checked
- Moved watched files since 016daa654: ui/hunter/presets.ts, ui/warrior/presets.ts, tools/dps_report/assumptions_def.py (plus orchestration/standards.md, not a watched path). The class branches show only stale copies of other classes' files (gotcha 1).
- hunter presets.ts: comments, the Marksmanship rotation text (Multi-Shot on the shared cooldown) and the Marksmanship race (Windshaper to Human). No Consumes, buff, debuff or gear change.
- warrior presets.ts: comments and Fury / Arms talent strings (flat Rage ruling 2026-10-25). No Consumes, buff, debuff or gear change.
- assumptions_def.py: hunter_tsa_hm_rank is now Trueshot Aura only (low = rank 4, +75); Hunter's Mark is settled at 143 (server hotfix), matching sim/core/debuffs.go ForeverHuntersMarkRank4 = 143. Cobra damage high setting and raptor8 row removed. No other buff value moved.
- standards.md: warrior Rage scaling ruling only.
- scan.sh: banned items, Revelation, trinkets, double crit elixir all empty. Minor-armor exclusive effects 3 (Dust Cloud, expected since pass 40). Air/weapon totems unchanged.
- Client: wago latest probe is 1.60.1.70532, same as state; no builddiff needed.

## Open items
None. Housekeeping: pass doc still says Hunter's Mark +140 and "75 / 98 mid, 50 / 71 low" for hunter_tsa_hm_rank; it is now 143 and a TSA-only row (low +75).
