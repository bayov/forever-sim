# Mechanics pass 89 (2026-11-18, master 8219f1083)

Range c198d81e3..8219f1083 (one commit). Client unchanged at 1.60.1.70959. Quiet: no new finding.
Open findings mechanics-20261118-1, mechanics-20261113-2 and mechanics-20261104-1 unchanged.

## Get current

| Source | Newest | New since pass 88 |
|---|---|---|
| Client build | 1.60.1.70959 (gate) | none |
| Dev notes t/2360696 | last post 2026-11-04 19:35Z | none |
| Hotfix post t/2403040 | post 12, 2026-11-18 03:24Z | none |
| Daily update / TASKS / forever-update merges | none in range | none |

## Commits

| Commit | Change | Verdict |
|---|---|---|
| 8219f1083 | Forever: cast speed hastes Hurricane and Volley ticks (sim/core/dot.go hastedTicks, HasteReducesDuration; druid/hurricane.go, hunter/volley.go) | User's ruling 2026-11-18 (wowsims/forever's rule), not a finding. Knock-on checked: only Hurricane and Volley set AffectedByCastSpeed under Forever (Mind Flay, Siphon Life, an item dot set false), so no other dot moves; Classic path unchanged (unrounded, tick count kept). Remaining ticks and the channel log include hasteTicks; SetTickPeriod (Wrack) is untouched. No report APL casts Hurricane or Volley, so no preset number moves. |

## Client check (70959, SpellMisc)

| Spell | Attributes_5 | Attributes_8 |
|---|---|---|
| Hurricane 16914 / 17401 / 17402 | 0x8000000 | 0x1000 |
| Volley 1510 / 14294 / 14295 | 0 | 0 |

No spell in the 70959 client carries Attributes_5 0x2000 (haste affects periodic) or Attributes_8 0x80000 (haste
affects duration), so the client data neither confirms nor contradicts the ruling. It stays a log check
(recorded under Settled in ref/mechanics/verdicts.md).
