# Fairness pass 44 (2026-11-16, master f30078d)

Result: quiet (no new finding). Range cfc7f93..f30078d.

## What changed in the watched paths

| Change | Commit | Effect on the comparison |
|---|---|---|
| Item database: 2026-11-15 data pull, Nov 13 Forever hotfixes, AH availability, Demondash Spaulders, Enh BiS list | d1cde9a5d, 4b3d8b7e9, e635633b2, 04e867105, f30078da1 | 823 db.json entries changed; none is an item or enchant used by any `tools/dps_report/req` request (`$AUDIT_WORK/fairness/dbdiff.py`) |
| `sim/core/sim_concurrent.go`: a player's own database added once before the split | 7429dc278 | concurrency / logcheck only; report requests carry no player database |
| standards 17: Windfury Weapon and Totem have separate cooldowns (user ruling 2026-11-15) | 85178ca55 | records existing behaviour; no request moved |
| Warrior default: two-hand Fury (`isDefault` flipped in the two Fury metas, ui/warrior/presets.ts) | aec7d79f1 | presets/metas only; the 1103.0 vs 1092.8 claim matches the sweep (1103.6 vs 1092.8) |
| Racial text from client 70938 | af62ee6da | ui/core text only |

## Checks

- Shared 20k run of f30078d (`/tmp/audit-work/runs/master`) vs `ref/fairness/sweep_baseline.log` (25ce81d, 5k) at
  each request's report race: 43 of 43 base builds within 2.1 (largest: Enh water-shield 1027.8 -> 1025.7, Demo
  1055.3 -> 1053.9, Frostfire 1065.7 -> 1064.3; all 5k-vs-20k noise, no code for those classes changed). No leak.
- Gear: `audit1b.py` 19 of 19 "5a PASS | 1b PASS".
- Published vs current (`runs/vs_published.txt`): the same meta lags as pass 43, unchanged in size: Affliction family
  +31.9 to +36.7, hunters -9.2 to -22.4, four rogues -10.5 to -20.1, Fire -4.9. Already filed by the results auditor
  (results-20261104-1 / -2 pattern); independently confirmed again, not refiled.
- Best race: unchanged from pass 43 (no build moved).

## Baseline

Not replaced: no build moved, so `baseline` stays 25ce81d. `orchestration/fairness_report.md` §1 got a one-line
"still current at f30078d" note.

## Open items (unchanged)

Hunter end-of-fight test, Ice Lance Low range, FTT + WF imbue stacking, boss partial resist log, Warrior Rage k and
demon/pet scaling ranges (see passes/fairness.md).
