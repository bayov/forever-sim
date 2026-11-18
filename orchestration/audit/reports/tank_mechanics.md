# Tank mechanics, pass 20

- Master: 8219f1083f (from c198d81e3a). Client: 1.60.1.70959, unchanged (wago latest, 2026-11-18).
- Result: quiet.

## What changed

| Commit | File | Tank effect |
|---|---|---|
| 8219f1083 | sim/druid/hurricane.go | None. Under Forever, cast speed now shortens Hurricane's tick length (`AffectedByCastSpeed`), keeping its 10 s duration. Hurricane is a Balance channel; no Feral tank preset, APL or `tools/tankbench` fight casts it (grep of ui tank/bear presets and tankbench: no hits). |

Other commits in the range are DPS Report records, a brief update and the audit loop's own commit; no tank
ability, talent or engine file changed.

## Engine, abilities, tests

Unchanged since pass 19; no re-check needed. Verdicts in `ref/tank_mechanics/verdicts.md` stand.

## Findings

No new findings. Open: 20261117-1, 20261112-2, 20261108-1, 20261107-1 (unchanged; none fixed by this range).
