# Tank mechanics, pass 12

Master 8306f1d52 (from a0b4d3be9). Client 1.60.1.70888, unchanged since pass 11.

## What changed

| Commit | File | Tank relevance |
|---|---|---|
| c380bbaef | sim/druid/rake.go | None. Rake's direct hit now adds 6.5% AP under Forever (user ruling 2026-11-12). Rake is a Cat Form ability; no bear preset, APL (`sim/druid/tank`, `ui/*tank*`) or tank test casts it. Bear code mentions Rake only in a Lacerate comment (Rend and Tear) and Savage Fury's spell list, neither of which this change touches. |

## Engine and specs

No engine, warrior, paladin, shaman or bear file changed. Verdicts in `ref/tank_mechanics/verdicts.md` stand.

## Findings

None new. Still open: 20261111-1, 20261108-1, 20261107-1 (their files unchanged on master).
