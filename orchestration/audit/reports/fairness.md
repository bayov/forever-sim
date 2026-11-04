# Fairness pass 28 (2026-11-04, master fb584ea)

Result: quiet (no new finding). 23 watched files changed since f331f0c (14 commits), but no `sim/`, `proto/` or
item database change. So the only builds that can move are the four Fury requests that 7b4dc5d97 changed. I re-swept
the two Fury base builds (5k, seed 1, every race) on the loop's fb584ea CLI and tree. The baseline stays f331f0c
(not a full sweep).

## What changed

| Change | Commit | Effect on fairness |
|---|---|---|
| Fury DW rotation (Bloodthirst first, HS from 60, Hamstring from 20, Bloodrage under 50) | 7b4dc5d97 | Fury DW 1153.6 -> 1152.5 at 5k (noise; worker's 20k: +0.4) |
| Fury 2H talents: Unbridled Wrath 5/5 for Furious Precision (2505001501...) | 7b4dc5d97 | Fury 2H 1170.7 -> 1172.3 at 5k (worker's 20k: +3.4) |
| Standards 38 (Forever only), 39 (cast time / 3.5 s coefficient) | d9ba4a55c | rulings only; Chain Lightning 0.714 not in code yet (shaman brief ca22004a3) |
| UI presets drop Salvation (`ui/*/presets.ts`) | fdb14acd0 | no request carries Salvation (grep), so no report number moves; threat only anyway |
| report_hashes.txt | fb584eafa | test data |

## What I checked

- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Shared 20k run (`runs/vs_published.txt`): 62 of 68 within 0.5. The six shamans (+199 to +277) are the stale metas in
  results-20261102-1 (not filed again; confirmed: no shaman input changed this window).
- Fury sweep (5k): DW Orc 1152.5 (published 1153.0, -0.5), Night Elf next at 1146.8; 2H Orc 1172.3 (published 1173.3,
  -1.0), Night Elf next at 1166.1. Report race is the best race on both. Gaps within the ±1.5 noise.
- Results overlap: results-20261102-1's 2026-11-04 update notes builds.json still has the pre-follow-up warrior numbers
  (waiting on the republish). Independent confirmation: the 5k sweep matches the new metas, not builds.json.

## Ranking

Unchanged: Lone Wolf 1314.4, MM 1254.6, BM 1195.7, Fury 2H 1172.3, Enhancement 1159.9 (published 883.5, stale),
Fury DW 1152.5, ... Smite 781.0 last. Top over bottom 1.68x. Full table: `orchestration/fairness_report.md` §1.

## Open items (carried)

- Next full sweep when `sim/` or the database changes (the shaman phase 4 port, Chain Lightning 0.714, will be one).
